package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/11DingKing/dual-teacher-practice-go/internal/clock"
	"github.com/11DingKing/dual-teacher-practice-go/internal/domain"
	"github.com/11DingKing/dual-teacher-practice-go/internal/repository"
	"github.com/11DingKing/dual-teacher-practice-go/internal/storage/sqlite"
)

func practiceFixture(t *testing.T) (Practice, *sqlite.DB) {
	t.Helper()
	db, e := sqlite.Open(context.Background(), ":memory:")
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { db.Close() })
	now := time.Date(2026, 8, 25, 0, 0, 0, 0, time.UTC)
	if _, e = db.SQL.Exec("INSERT INTO users VALUES('u','u','p','teacher','T',NULL,'2026-01-01')"); e != nil {
		t.Fatal(e)
	}
	if _, e = db.SQL.Exec("INSERT INTO quotas VALUES('q','Q',100,0,1000,0,1,1)"); e != nil {
		t.Fatal(e)
	}
	if _, e = db.SQL.Exec("INSERT INTO applications VALUES('a','u','q',2026,'college_approved',10,10,'a',1,'2026-12-01','2026-01-01','2026-01-01')"); e != nil {
		t.Fatal(e)
	}
	p := domain.Practice{ID: "p", ApplicationID: "a", CompanyName: "企业", MentorName: "导师", ContactEmail: "e@example.com", Status: "planned", PlannedHours: 100, StartsAt: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC), EndsAt: time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC), Version: 1}
	if e = (repository.Practices{DB: db.SQL}).Add(context.Background(), p); e != nil {
		t.Fatal(e)
	}
	return Practice{Practices: repository.Practices{DB: db.SQL}, Apps: repository.Applications{DB: db.SQL}, Audit: repository.Audit{DB: db.SQL}, Clock: clock.Fixed{T: now}}, db
}

func assertPracticeUntouched(t *testing.T, db *sqlite.DB) {
	t.Helper()
	var actual, version int
	if e := db.SQL.QueryRow("SELECT actual_hours,version FROM practices WHERE id='p'").Scan(&actual, &version); e != nil {
		t.Fatal(e)
	}
	if actual != 0 || version != 1 {
		t.Fatalf("practice data mutated despite cancelled/timeout: actual=%d version=%d", actual, version)
	}
	var n int
	if e := db.SQL.QueryRow("SELECT COUNT(*) FROM audit_events WHERE action='practice_hours_recorded'").Scan(&n); e != nil || n != 0 {
		t.Fatalf("audit written despite cancelled/timeout: %d %v", n, e)
	}
}

// A request whose context is already cancelled before the service runs must not
// write any practice hours or audit row.
func TestRecordHoursCancelledContextDoesNotMutate(t *testing.T) {
	s, db := practiceFixture(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if e := s.RecordHours(ctx, "p", 40, "u", "req"); e == nil {
		t.Fatal("cancelled request accepted")
	}
	assertPracticeUntouched(t, db)
}

// A request that times out mid-flight must roll back the whole transaction:
// neither the hours row nor the audit event may be persisted.
func TestRecordHoursTimeoutContextDoesNotMutate(t *testing.T) {
	s, db := practiceFixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(2 * time.Millisecond) // let the deadline elapse before the call

	e := s.RecordHours(ctx, "p", 40, "u", "req")
	if e == nil {
		t.Fatal("timed-out request accepted")
	}
	assertPracticeUntouched(t, db)
}

// The happy path still commits hours and audit atomically.
func TestRecordHoursCommitsAtomically(t *testing.T) {
	s, db := practiceFixture(t)
	if e := s.RecordHours(context.Background(), "p", 40, "u", "req"); e != nil {
		t.Fatal(e)
	}
	var actual, version int
	if e := db.SQL.QueryRow("SELECT actual_hours,version FROM practices WHERE id='p'").Scan(&actual, &version); e != nil {
		t.Fatal(e)
	}
	if actual != 40 || version != 2 {
		t.Fatalf("actual=%d version=%d", actual, version)
	}
	var n int
	if e := db.SQL.QueryRow("SELECT COUNT(*) FROM audit_events WHERE action='practice_hours_recorded'").Scan(&n); e != nil || n != 1 {
		t.Fatalf("audit=%d %v", n, e)
	}
}

func TestRecordHoursRejectsOverPlanned(t *testing.T) {
	s, db := practiceFixture(t)
	if e := s.RecordHours(context.Background(), "p", 101, "u", "req"); !errors.Is(e, domain.ErrQuotaExceeded) {
		t.Fatalf("%v", e)
	}
	assertPracticeUntouched(t, db)
}
