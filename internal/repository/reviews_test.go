package repository

import (
	"context"
	"github.com/11DingKing/dual-teacher-practice-go/internal/domain"
	"github.com/11DingKing/dual-teacher-practice-go/internal/storage/sqlite"
	"testing"
	"time"
)

func reviewDB(t *testing.T) *sqlite.DB {
	db := practiceDB(t)
	if _, e := db.SQL.Exec("UPDATE applications SET status='peer_review' WHERE id='a'"); e != nil {
		t.Fatal(e)
	}
	return db
}
func TestReviewAddAndList(t *testing.T) {
	db := reviewDB(t)
	r := Reviews{DB: db.SQL}
	now := time.Now().UTC()
	items := []domain.Review{{ID: "r1", ApplicationID: "a", ReviewerID: "u", ReviewerRole: domain.Role学院, Score: 80, Decision: domain.ReviewPass, Comment: "通过", CreatedAt: now}, {ID: "r2", ApplicationID: "a", ReviewerID: "u2", ReviewerRole: domain.RoleEnterprise, Score: 70, Decision: domain.ReviewPass, Comment: "通过", CreatedAt: now.Add(time.Minute)}}
	for _, v := range items {
		if e := r.Add(context.Background(), v); e != nil {
			t.Fatal(e)
		}
	}
	got, e := r.ForApplication(context.Background(), "a")
	if e != nil || len(got) != 2 {
		t.Fatalf("%+v %v", got, e)
	}
	if got[0].ReviewerRole != domain.Role学院 || got[1].Score != 70 {
		t.Fatalf("%+v", got)
	}
}
func TestReviewUniqueReviewer(t *testing.T) {
	db := reviewDB(t)
	r := Reviews{DB: db.SQL}
	v := domain.Review{ID: "r1", ApplicationID: "a", ReviewerID: "u", ReviewerRole: domain.Role学院, Score: 80, Decision: domain.ReviewPass, Comment: "ok", CreatedAt: time.Now().UTC()}
	if e := r.Add(context.Background(), v); e != nil {
		t.Fatal(e)
	}
	v.ID = "r2"
	if e := r.Add(context.Background(), v); e == nil {
		t.Fatal("duplicate reviewer accepted")
	}
}
func TestReviewMissingApplication(t *testing.T) {
	db := reviewDB(t)
	r := Reviews{DB: db.SQL}
	v := domain.Review{ID: "r", ApplicationID: "missing", ReviewerID: "u", ReviewerRole: domain.Role学院, Score: 80, Decision: domain.ReviewPass, Comment: "ok", CreatedAt: time.Now().UTC()}
	if e := r.Add(context.Background(), v); e == nil {
		t.Fatal("orphan review accepted")
	}
}
