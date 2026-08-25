package service

import (
	"context"
	"github.com/11DingKing/dual-teacher-practice-go/internal/clock"
	"github.com/11DingKing/dual-teacher-practice-go/internal/domain"
	"github.com/11DingKing/dual-teacher-practice-go/internal/repository"
	"github.com/11DingKing/dual-teacher-practice-go/internal/storage/sqlite"
	"sync"
	"testing"
	"time"
)

func appFixture(t *testing.T) (Application, *sqlite.DB) {
	t.Helper()
	db, e := sqlite.Open(context.Background(), ":memory:")
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { db.Close() })
	now := time.Date(2026, 8, 25, 0, 0, 0, 0, time.UTC)
	users := repository.Users{DB: db.SQL}
	if e = users.Create(context.Background(), domain.User{ID: "u", Username: "teacher", Role: domain.RoleTeacher, DisplayName: "教师", CreatedAt: now}, "pw"); e != nil {
		t.Fatal(e)
	}
	if e = (repository.Quotas{DB: db.SQL}).Create(context.Background(), domain.Quota{ID: "q", Name: "年度", MaxHours: 100, MaxBudgetCents: 100000, Version: 1, Active: true}); e != nil {
		t.Fatal(e)
	}
	return Application{DB: db.SQL, Apps: repository.Applications{DB: db.SQL}, Quotas: repository.Quotas{DB: db.SQL}, Audit: repository.Audit{DB: db.SQL}, Clock: clock.Fixed{T: now}}, db
}
func TestSubmitReservesQuotaAndAudits(t *testing.T) {
	s, db := appFixture(t)
	a, e := s.Submit(context.Background(), domain.Application{ID: "a", TeacherID: "u", QuotaID: "q", Year: 2026, RequestedHours: 20, RequestedBudgetCents: 1000, IdempotencyKey: "key"}, "u", "req")
	if e != nil {
		t.Fatal(e)
	}
	if a.Status != domain.StatusDraft {
		t.Fatal(a.Status)
	}
	var h, b int
	if e = db.SQL.QueryRow("SELECT used_hours,used_budget_cents FROM quotas WHERE id='q'").Scan(&h, &b); e != nil {
		t.Fatal(e)
	}
	if h != 20 || b != 1000 {
		t.Fatalf("%d %d", h, b)
	}
	var n int
	if e = db.SQL.QueryRow("SELECT COUNT(*) FROM audit_events").Scan(&n); e != nil || n != 1 {
		t.Fatalf("%d %v", n, e)
	}
}
func TestSubmitRejectsQuotaOverflow(t *testing.T) {
	s, db := appFixture(t)
	if _, e := s.Submit(context.Background(), domain.Application{ID: "a", TeacherID: "u", QuotaID: "q", Year: 2026, RequestedHours: 101, RequestedBudgetCents: 1, IdempotencyKey: "k"}, "u", "r"); e != domain.ErrQuotaExceeded {
		t.Fatalf("%v", e)
	}
	var n int
	_ = db.SQL.QueryRow("SELECT COUNT(*) FROM applications").Scan(&n)
	if n != 0 {
		t.Fatal(n)
	}
}
func TestSubmitRequiresPositiveHours(t *testing.T) {
	s, _ := appFixture(t)
	for _, h := range []int{0, -1} {
		if _, e := s.Submit(context.Background(), domain.Application{ID: token(), TeacherID: "u", QuotaID: "q", Year: 2026 + h, RequestedHours: h, RequestedBudgetCents: 1, IdempotencyKey: token()}, "u", "r"); e == nil {
			t.Fatalf("hours %d accepted", h)
		}
	}
}
func TestTransitionChecksState(t *testing.T) {
	s, _ := appFixture(t)
	_, e := s.Submit(context.Background(), domain.Application{ID: "a", TeacherID: "u", QuotaID: "q", Year: 2026, RequestedHours: 1, RequestedBudgetCents: 1, IdempotencyKey: "k"}, "u", "r")
	if e != nil {
		t.Fatal(e)
	}
	if e = s.Transition(context.Background(), "a", domain.StatusCertified, "u", "r"); e == nil {
		t.Fatal("invalid transition accepted")
	}
}
func TestTransitionMovesState(t *testing.T) {
	s, _ := appFixture(t)
	_, e := s.Submit(context.Background(), domain.Application{ID: "a", TeacherID: "u", QuotaID: "q", Year: 2026, RequestedHours: 1, RequestedBudgetCents: 1, IdempotencyKey: "k"}, "u", "r")
	if e != nil {
		t.Fatal(e)
	}
	if e = s.Transition(context.Background(), "a", domain.StatusSubmitted, "u", "r"); e != nil {
		t.Fatal(e)
	}
	a, e := s.Get(context.Background(), "a")
	if e != nil || a.Status != domain.StatusSubmitted {
		t.Fatalf("%+v %v", a, e)
	}
}
func TestConcurrentQuotaReservations(t *testing.T) {
	s, _ := appFixture(t)
	ctx := context.Background()
	var wg sync.WaitGroup
	success := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, e := s.Submit(ctx, domain.Application{ID: token(), TeacherID: "u", QuotaID: "q", Year: 2026 + i, RequestedHours: 20, RequestedBudgetCents: 1, IdempotencyKey: token()}, "u", "r")
			success <- e == nil
		}(i)
	}
	wg.Wait()
	close(success)
	count := 0
	for ok := range success {
		if ok {
			count++
		}
	}
	if count > 5 {
		t.Fatalf("quota overcommitted: %d", count)
	}
}

// A failed operation must not leave committed business state behind. The audit
// record is written inside the submit transaction; when the actor id is
// invalidated (operator info no longer resolvable) the audit INSERT fails on
// its actor_id foreign key, the whole transaction rolls back, and neither the
// application row nor the reserved quota may survive.
func TestSubmitAuditFailureLeavesNoResidue(t *testing.T) {
	s, db := appFixture(t)
	_, e := s.Submit(context.Background(), domain.Application{ID: "a", TeacherID: "u", QuotaID: "q", Year: 2026, RequestedHours: 20, RequestedBudgetCents: 1000, IdempotencyKey: "key"}, "invalidated-actor", "req")
	if e == nil {
		t.Fatal("expected audit failure error, got nil")
	}
	var appN, auditN, usedH, usedB int
	_ = db.SQL.QueryRow("SELECT COUNT(*) FROM applications").Scan(&appN)
	_ = db.SQL.QueryRow("SELECT COUNT(*) FROM audit_events").Scan(&auditN)
	_ = db.SQL.QueryRow("SELECT used_hours, used_budget_cents FROM quotas WHERE id='q'").Scan(&usedH, &usedB)
	if appN != 0 || auditN != 0 || usedH != 0 || usedB != 0 {
		t.Fatalf("failed op left residue: apps=%d audit=%d used_hours=%d used_budget=%d", appN, auditN, usedH, usedB)
	}
}
