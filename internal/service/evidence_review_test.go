package service

import (
	"context"
	"github.com/11DingKing/dual-teacher-practice-go/internal/clock"
	"github.com/11DingKing/dual-teacher-practice-go/internal/domain"
	"github.com/11DingKing/dual-teacher-practice-go/internal/repository"
	"github.com/11DingKing/dual-teacher-practice-go/internal/storage/sqlite"
	"testing"
	"time"
)

func evidenceFixture(t *testing.T) (Evidence, Review, *sqlite.DB) {
	db := mustServiceDB(t)
	now := time.Date(2026, 8, 25, 0, 0, 0, 0, time.UTC)
	if _, e := db.SQL.Exec("INSERT INTO users VALUES('u','u','p','teacher','T',NULL,'2026-01-01'),('r1','r1','p','college','R1',NULL,'2026-01-01'),('r2','r2','p','enterprise','R2',NULL,'2026-01-01')"); e != nil {
		t.Fatal(e)
	}
	if _, e := db.SQL.Exec("INSERT INTO quotas VALUES('q','Q',100,0,10000,0,1,1)"); e != nil {
		t.Fatal(e)
	}
	if _, e := db.SQL.Exec("INSERT INTO applications VALUES('a','u','q',2026,'evidence_review',10,10,'a',1,'2026-12-01','2026-01-01','2026-01-01')"); e != nil {
		t.Fatal(e)
	}
	a := repository.Applications{DB: db.SQL}
	return Evidence{Evidence: repository.Evidences{DB: db.SQL}, Apps: a, Audit: repository.Audit{DB: db.SQL}, Clock: clock.Fixed{T: now}}, Review{Reviews: repository.Reviews{DB: db.SQL}, Evidences: repository.Evidences{DB: db.SQL}, Apps: a, Audit: repository.Audit{DB: db.SQL}, Clock: clock.Fixed{T: now}}, db
}
func mustServiceDB(t *testing.T) *sqlite.DB {
	db, e := sqlite.Open(context.Background(), ":memory:")
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { db.Close() })
	return db
}
func TestEvidenceServiceAddAndVerify(t *testing.T) {
	e, _, db := evidenceFixture(t)
	v := domain.Evidence{ID: "e", ApplicationID: "a", Kind: "practice_log", Title: "实践日志", URI: "https://e"}
	if x := e.Add(context.Background(), v, "u", "req"); x != nil {
		t.Fatal(x)
	}
	if x := e.Verify(context.Background(), "e", "r1", "req", true, "已核验"); x != nil {
		t.Fatal(x)
	}
	var status string
	if x := db.SQL.QueryRow("SELECT verification_status FROM evidences WHERE id='e'").Scan(&status); x != nil || status != "verified" {
		t.Fatalf("%s %v", status, x)
	}
}
func TestEvidenceServiceRejectsWrongWindow(t *testing.T) {
	e, _, db := evidenceFixture(t)
	if _, x := db.SQL.Exec("UPDATE applications SET status='submitted' WHERE id='a'"); x != nil {
		t.Fatal(x)
	}
	if x := e.Add(context.Background(), domain.Evidence{ID: "e", ApplicationID: "a", Kind: "k", Title: "t", URI: "u"}, "u", "r"); x == nil {
		t.Fatal("wrong window accepted")
	}
}
func TestEvidenceServiceCorrection(t *testing.T) {
	e, _, _ := evidenceFixture(t)
	if x := e.Add(context.Background(), domain.Evidence{ID: "e", ApplicationID: "a", Kind: "k", Title: "t", URI: "u"}, "u", "r"); x != nil {
		t.Fatal(x)
	}
	if x := e.Verify(context.Background(), "e", "r1", "r", false, "请补交原件"); x != nil {
		t.Fatal(x)
	}
	items, x := e.Evidence.ForApplication(context.Background(), "a")
	if x != nil || items[0].Status != domain.EvidenceCorrection {
		t.Fatalf("%+v %v", items, x)
	}
}
func TestReviewServiceNeedsPeerState(t *testing.T) {
	_, r, _ := evidenceFixture(t)
	if x := r.Add(context.Background(), domain.Review{ID: "r", ApplicationID: "a", Score: 80, Decision: domain.ReviewPass, Comment: "好"}, "r1", "req"); x == nil {
		t.Fatal("review accepted before peer state")
	}
}
func TestReviewServiceScoreBounds(t *testing.T) {
	e, r, db := evidenceFixture(t)
	_, _ = e, db
	_, x := db.SQL.Exec("UPDATE applications SET status='peer_review' WHERE id='a'")
	if x != nil {
		t.Fatal(x)
	}
	for i, score := range []int{-1, 101} {
		if x = r.Add(context.Background(), domain.Review{ID: token(), ApplicationID: "a", Score: score, Decision: domain.ReviewPass, Comment: "x"}, "r1", "req"); x == nil {
			t.Fatalf("score %d accepted", i)
		}
	}
}
func TestReviewServiceStoresRoleAndDecision(t *testing.T) {
	_, r, db := evidenceFixture(t)
	_, x := db.SQL.Exec("UPDATE applications SET status='peer_review' WHERE id='a'")
	if x != nil {
		t.Fatal(x)
	}
	if x = r.Add(context.Background(), domain.Review{ID: "r", ApplicationID: "a", Score: 88, Decision: domain.ReviewPass, Comment: "通过", ReviewerRole: domain.Role学院}, "r1", "req"); x != nil {
		t.Fatal(x)
	}
	var role, decision string
	if x = db.SQL.QueryRow("SELECT reviewer_role,decision FROM reviews WHERE id='r'").Scan(&role, &decision); x != nil || role != "college" || decision != "pass" {
		t.Fatalf("%s %s %v", role, decision, x)
	}
}
