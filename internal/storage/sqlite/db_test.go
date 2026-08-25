package sqlite

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
)

func openTestDB(t *testing.T) *DB {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	db, e := Open(context.Background(), path)
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { db.Close() })
	return db
}
func TestMigrationCreatesRelations(t *testing.T) {
	db := openTestDB(t)
	tables := []string{"users", "sessions", "quotas", "applications", "practices", "evidences", "reviews", "audit_events", "idempotency_keys", "worker_jobs"}
	for _, name := range tables {
		var got string
		if e := db.SQL.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", name).Scan(&got); e != nil {
			t.Fatalf("%s missing: %v", name, e)
		}
	}
}
func TestMigrationIsIdempotent(t *testing.T) {
	db := openTestDB(t)
	if e := db.Migrate(context.Background()); e != nil {
		t.Fatal(e)
	}
	var n int
	if e := db.SQL.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&n); e != nil {
		t.Fatal(e)
	}
	if n != 1 {
		t.Fatalf("migrations=%d", n)
	}
}
func TestForeignKeyRejectsOrphan(t *testing.T) {
	db := openTestDB(t)
	_, e := db.SQL.Exec("INSERT INTO sessions(id,user_id,expires_at,created_at) VALUES('s','missing','2026-01-01','2026-01-01')")
	if e == nil {
		t.Fatal("expected foreign key error")
	}
}
func TestUniqueApplicationYear(t *testing.T) {
	db := openTestDB(t)
	_, e := db.SQL.Exec("INSERT INTO users VALUES('u','u','p','teacher','T',NULL,'2026-01-01')")
	if e != nil {
		t.Fatal(e)
	}
	_, e = db.SQL.Exec("INSERT INTO quotas VALUES('q','Q',100,0,1000,0,1,1)")
	if e != nil {
		t.Fatal(e)
	}
	args := []any{"a", "u", "q", 2026, "draft", 10, 10, "i", 1, "2026-12-01", "2026-01-01", "2026-01-01"}
	if _, e = db.SQL.Exec("INSERT INTO applications VALUES(?,?,?,?,?,?,?,?,?,?,?,?)", args...); e != nil {
		t.Fatal(e)
	}
	args[0] = "b"
	args[7] = "j"
	if _, e = db.SQL.Exec("INSERT INTO applications VALUES(?,?,?,?,?,?,?,?,?,?,?,?)", args...); e == nil {
		t.Fatal("expected unique year error")
	}
}
func TestRollbackLeavesNoRows(t *testing.T) {
	db := openTestDB(t)
	tx, e := db.SQL.Begin()
	if e != nil {
		t.Fatal(e)
	}
	if _, e = tx.Exec("INSERT INTO users VALUES('u','u','p','teacher','T',NULL,'2026-01-01')"); e != nil {
		t.Fatal(e)
	}
	if e = tx.Rollback(); e != nil {
		t.Fatal(e)
	}
	var n int
	if e = db.SQL.QueryRow("SELECT COUNT(*) FROM users").Scan(&n); e != nil && e != sql.ErrNoRows {
		t.Fatal(e)
	}
	if n != 0 {
		t.Fatalf("rows=%d", n)
	}
}
