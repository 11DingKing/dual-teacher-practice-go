package repository

import (
	"context"
	"database/sql"
	"github.com/11DingKing/dual-teacher-practice-go/internal/domain"
	"time"
)

type Idempotency struct{ DB *sql.DB }

func (r Idempotency) Lookup(ctx context.Context, key, actor, operation string, now time.Time) (string, error) {
	var response, expires string
	e := r.DB.QueryRowContext(ctx, "SELECT response,expires_at FROM idempotency_keys WHERE key=? AND actor_id=? AND operation=?", key, actor, operation).Scan(&response, &expires)
	if e != nil {
		return "", domain.ErrNotFound
	}
	deadline, _ := time.Parse(time.RFC3339Nano, expires)
	if !deadline.After(now) {
		return "", domain.ErrNotFound
	}
	return response, nil
}
func (r Idempotency) Save(ctx context.Context, key, actor, operation, response string, now time.Time) error {
	_, e := r.DB.ExecContext(ctx, "INSERT INTO idempotency_keys(key,actor_id,operation,response,created_at,expires_at) VALUES(?,?,?,?,?,?)", key, actor, operation, response, now.Format(time.RFC3339Nano), now.Add(24*time.Hour).Format(time.RFC3339Nano))
	return e
}
func (r Idempotency) Purge(ctx context.Context, now time.Time) (int64, error) {
	res, e := r.DB.ExecContext(ctx, "DELETE FROM idempotency_keys WHERE expires_at<=?", now.Format(time.RFC3339Nano))
	if e != nil {
		return 0, e
	}
	return res.RowsAffected()
}
