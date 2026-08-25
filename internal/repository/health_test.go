package repository

import (
	"context"
	"github.com/11DingKing/dual-teacher-practice-go/internal/storage/sqlite"
	"testing"
)

func TestHealthReadyAndTableCount(t *testing.T) {
	db, e := sqlite.Open(context.Background(), ":memory:")
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	h := Health{DB: db.SQL}
	if e = h.Ready(context.Background()); e != nil {
		t.Fatal(e)
	}
	n, e := h.TableCount(context.Background())
	if e != nil || n < 11 {
		t.Fatalf("%d %v", n, e)
	}
}
func TestHealthCancelledContext(t *testing.T) {
	db, e := sqlite.Open(context.Background(), ":memory:")
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if e = (Health{DB: db.SQL}).Ready(ctx); e == nil {
		t.Fatal("cancelled health succeeded")
	}
}
