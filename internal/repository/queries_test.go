package repository

import (
	"context"
	"github.com/11DingKing/dual-teacher-practice-go/internal/domain"
	"github.com/11DingKing/dual-teacher-practice-go/internal/storage/sqlite"
	"testing"
	"time"
)

func queryDB(t *testing.T) *sqlite.DB {
	db, e := sqlite.Open(context.Background(), ":memory:")
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { db.Close() })
	now := "2026-01-01T00:00:00Z"
	_, e = db.SQL.Exec("INSERT INTO users VALUES('u','u','p','teacher','T',NULL,?)", now)
	if e != nil {
		t.Fatal(e)
	}
	_, e = db.SQL.Exec("INSERT INTO quotas VALUES('q','Q',100,0,1000,0,1,1)")
	if e != nil {
		t.Fatal(e)
	}
	for i := 0; i < 4; i++ {
		id := "a" + string(rune('0'+i))
		status := "submitted"
		if i == 3 {
			status = "certified"
		}
		_, e = db.SQL.Exec("INSERT INTO applications VALUES(?,?,?,?,?,?,?,?,?,?,?,?)", id, "u", "q", 2026+i, status, 10, 10, id, 1, "2026-01-01T00:00:00Z", now, now)
		if e != nil {
			t.Fatal(e)
		}
	}
	return db
}
func TestDueApplicationsFiltersTerminal(t *testing.T) {
	db := queryDB(t)
	r := Queries{DB: db.SQL}
	items, e := r.DueApplications(context.Background(), time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC), 10)
	if e != nil {
		t.Fatal(e)
	}
	if len(items) != 3 {
		t.Fatalf("%d", len(items))
	}
	for _, a := range items {
		if a.Status == domain.StatusCertified || a.Status == domain.StatusRejected || a.Status == domain.StatusExpired {
			t.Fatal(a.Status)
		}
	}
}
func TestDueApplicationsLimit(t *testing.T) {
	db := queryDB(t)
	r := Queries{DB: db.SQL}
	for _, limit := range []int{1, 2, 3, 10} {
		items, e := r.DueApplications(context.Background(), time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC), limit)
		if e != nil {
			t.Fatal(e)
		}
		want := limit
		if want > 3 {
			want = 3
		}
		if len(items) != want {
			t.Fatalf("limit %d got %d", limit, len(items))
		}
	}
}
func TestEvidenceSummary(t *testing.T) {
	db := queryDB(t)
	_, e := db.SQL.Exec("INSERT INTO evidences(id,application_id,kind,title,uri,verification_status,submitted_by,created_at) VALUES('e1','a0','k','t','u','verified','u','2026-01-01') ,('e2','a0','k','t','u','pending','u','2026-01-01'),('e3','a0','k','t','u','verified','u','2026-01-01')")
	if e != nil {
		t.Fatal(e)
	}
	m, e := (Queries{DB: db.SQL}).EvidenceSummary(context.Background(), "a0")
	if e != nil {
		t.Fatal(e)
	}
	if m["verified"] != 2 || m["pending"] != 1 {
		t.Fatalf("%v", m)
	}
}
func TestEvidenceSummaryEmpty(t *testing.T) {
	db := queryDB(t)
	m, e := (Queries{DB: db.SQL}).EvidenceSummary(context.Background(), "missing")
	if e != nil {
		t.Fatal(e)
	}
	if len(m) != 0 {
		t.Fatal(m)
	}
}
