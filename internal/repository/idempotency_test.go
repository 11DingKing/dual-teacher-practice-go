package repository

import (
	"context"
	"github.com/11DingKing/dual-teacher-practice-go/internal/storage/sqlite"
	"testing"
	"time"
)

func idemDB(t *testing.T) *sqlite.DB {
	db, e := sqlite.Open(context.Background(), ":memory:")
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { db.Close() })
	return db
}
func TestIdempotencySaveLookup(t *testing.T) {
	db := idemDB(t)
	r := Idempotency{DB: db.SQL}
	now := time.Date(2026, 8, 25, 0, 0, 0, 0, time.UTC)
	if e := r.Save(context.Background(), "k", "u", "submit", "{\"id\":1}", now); e != nil {
		t.Fatal(e)
	}
	v, e := r.Lookup(context.Background(), "k", "u", "submit", now.Add(time.Hour))
	if e != nil || v != "{\"id\":1}" {
		t.Fatalf("%q %v", v, e)
	}
	if _, e = r.Lookup(context.Background(), "k", "other", "submit", now); e == nil {
		t.Fatal("actor isolation failed")
	}
	if _, e = r.Lookup(context.Background(), "k", "u", "other", now); e == nil {
		t.Fatal("operation isolation failed")
	}
}
func TestIdempotencyExpiry(t *testing.T) {
	db := idemDB(t)
	r := Idempotency{DB: db.SQL}
	now := time.Date(2026, 8, 25, 0, 0, 0, 0, time.UTC)
	if e := r.Save(context.Background(), "k", "u", "op", "v", now); e != nil {
		t.Fatal(e)
	}
	if _, e := r.Lookup(context.Background(), "k", "u", "op", now.Add(25*time.Hour)); e == nil {
		t.Fatal("expired value returned")
	}
	n, e := r.Purge(context.Background(), now.Add(25*time.Hour))
	if e != nil || n != 1 {
		t.Fatalf("%d %v", n, e)
	}
}
func TestIdempotencyDuplicateRejected(t *testing.T) {
	db := idemDB(t)
	r := Idempotency{DB: db.SQL}
	now := time.Now().UTC()
	if e := r.Save(context.Background(), "k", "u", "op", "one", now); e != nil {
		t.Fatal(e)
	}
	if e := r.Save(context.Background(), "k", "u", "op", "two", now); e == nil {
		t.Fatal("duplicate accepted")
	}
}
