package repository

import (
	"context"
	"github.com/11DingKing/dual-teacher-practice-go/internal/domain"
	"github.com/11DingKing/dual-teacher-practice-go/internal/storage/sqlite"
	"testing"
	"time"
)

func practiceDB(t *testing.T) *sqlite.DB {
	db, e := sqlite.Open(context.Background(), ":memory:")
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { db.Close() })
	_, e = db.SQL.Exec("INSERT INTO users VALUES('u','u','p','teacher','T',NULL,'2026-01-01'),('u2','u2','p','enterprise','E',NULL,'2026-01-01')")
	if e != nil {
		t.Fatal(e)
	}
	_, e = db.SQL.Exec("INSERT INTO quotas VALUES('q','Q',100,0,1000,0,1,1)")
	if e != nil {
		t.Fatal(e)
	}
	_, e = db.SQL.Exec("INSERT INTO applications VALUES('a','u','q',2026,'college_approved',10,10,'a',1,'2026-12-01','2026-01-01','2026-01-01')")
	if e != nil {
		t.Fatal(e)
	}
	return db
}
func TestPracticeAddGet(t *testing.T) {
	db := practiceDB(t)
	r := Practices{DB: db.SQL}
	p := domain.Practice{ID: "p", ApplicationID: "a", CompanyName: "企业", MentorName: "导师", ContactEmail: "a@example.com", Status: "planned", PlannedHours: 100, StartsAt: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC), EndsAt: time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC), Version: 1}
	if e := r.Add(context.Background(), p); e != nil {
		t.Fatal(e)
	}
	got, e := r.Get(context.Background(), "p")
	if e != nil || got.CompanyName != p.CompanyName || got.Version != 1 {
		t.Fatalf("%+v %v", got, e)
	}
}
func TestPracticeOptimisticVersion(t *testing.T) {
	db := practiceDB(t)
	r := Practices{DB: db.SQL}
	p := domain.Practice{ID: "p", ApplicationID: "a", CompanyName: "企业", MentorName: "导师", ContactEmail: "a@example.com", Status: "planned", PlannedHours: 100, StartsAt: time.Now().UTC(), EndsAt: time.Now().UTC().Add(time.Hour), Version: 1}
	if e := r.Add(context.Background(), p); e != nil {
		t.Fatal(e)
	}
	if e := r.UpdateHours(context.Background(), "p", 20, 1); e != nil {
		t.Fatal(e)
	}
	if e := r.UpdateHours(context.Background(), "p", 30, 1); e != domain.ErrConflict {
		t.Fatalf("%v", e)
	}
	got, e := r.Get(context.Background(), "p")
	if e != nil || got.ActualHours != 20 || got.Version != 2 {
		t.Fatalf("%+v %v", got, e)
	}
}
func TestPracticeForeignKey(t *testing.T) {
	db := practiceDB(t)
	r := Practices{DB: db.SQL}
	p := domain.Practice{ID: "bad", ApplicationID: "missing", CompanyName: "企业", MentorName: "导师", ContactEmail: "a@example.com", Status: "planned", PlannedHours: 10, StartsAt: time.Now().UTC(), EndsAt: time.Now().UTC().Add(time.Hour), Version: 1}
	if e := r.Add(context.Background(), p); e == nil {
		t.Fatal("orphan practice accepted")
	}
}
