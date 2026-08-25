package repository

import (
	"context"
	"database/sql"
	"errors"
	"github.com/11DingKing/dual-teacher-practice-go/internal/storage/sqlite"
	"testing"
	"time"
)

func TestTransactionalCommitAndRollback(t *testing.T) {
	db, e := sqlite.Open(context.Background(), ":memory:")
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	r := Transactional{DB: db.SQL}
	if e = r.With(context.Background(), func(tx *sql.Tx) error { return nil }); e != nil {
		t.Fatal(e)
	}
	if e = r.With(context.Background(), func(tx *sql.Tx) error { return errors.New("stop") }); e == nil {
		t.Fatal("rollback callback accepted")
	}
	if e = r.Health(context.Background()); e != nil {
		t.Fatal(e)
	}
}
func TestBackoffMonotonic(t *testing.T) {
	prev := time.Duration(0)
	for i := 1; i <= 6; i++ {
		got := Backoff(i)
		if got <= prev {
			t.Fatalf("%d %v", i, got)
		}
		prev = got
	}
	if Backoff(0) != 100*time.Millisecond {
		t.Fatal(Backoff(0))
	}
	if Backoff(7) != 3200*time.Millisecond {
		t.Fatal(Backoff(7))
	}
}
func TestRetryableErrors(t *testing.T) {
	for _, msg := range []string{"database is locked", "busy"} {
		if !Retryable(errors.New(msg)) {
			t.Fatal(msg)
		}
	}
	if Retryable(errors.New("other")) {
		t.Fatal("other marked retryable")
	}
	if Retryable(nil) {
		t.Fatal("nil retryable")
	}
}
