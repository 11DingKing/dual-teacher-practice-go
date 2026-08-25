package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Transactional struct{ DB *sql.DB }

func (r Transactional) With(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, e := r.DB.BeginTx(ctx, nil)
	if e != nil {
		return fmt.Errorf("begin: %w", e)
	}
	if e = fn(tx); e != nil {
		_ = tx.Rollback()
		return e
	}
	if e = tx.Commit(); e != nil {
		return fmt.Errorf("commit: %w", e)
	}
	return nil
}
func (r Transactional) Health(ctx context.Context) error {
	var one int
	if e := r.DB.QueryRowContext(ctx, "SELECT 1").Scan(&one); e != nil {
		return e
	}
	if one != 1 {
		return fmt.Errorf("unexpected health value")
	}
	return nil
}
func Retryable(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return s == "database is locked" || s == "busy"
}
func Backoff(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	if attempt > 6 {
		attempt = 6
	}
	return time.Duration(1<<uint(attempt-1)) * 100 * time.Millisecond
}
