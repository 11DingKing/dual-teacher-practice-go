package repository

import (
	"context"
	"database/sql"
	"time"
)

type Health struct{ DB *sql.DB }

func (h Health) Ready(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	var value int
	return h.DB.QueryRowContext(ctx, "SELECT 1").Scan(&value)
}
func (h Health) TableCount(ctx context.Context) (int, error) {
	var n int
	e := h.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM sqlite_master WHERE type='table'").Scan(&n)
	return n, e
}
