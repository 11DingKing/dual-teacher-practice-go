package repository

import (
	"context"
	"database/sql"
	"github.com/11DingKing/dual-teacher-practice-go/internal/domain"
	"time"
)

type Sessions struct{ DB *sql.DB }

func (r Sessions) Create(ctx context.Context, s domain.Session) error {
	_, e := r.DB.ExecContext(ctx, "INSERT INTO sessions(id,user_id,expires_at,created_at) VALUES(?,?,?,?)", s.ID, s.UserID, s.ExpiresAt.Format(time.RFC3339Nano), s.CreatedAt.Format(time.RFC3339Nano))
	return e
}
func (r Sessions) GetActive(ctx context.Context, id string, now time.Time) (domain.Session, error) {
	var s domain.Session
	var exp, created string
	var rev sql.NullString
	e := r.DB.QueryRowContext(ctx, "SELECT id,user_id,expires_at,created_at,revoked_at FROM sessions WHERE id=?", id).Scan(&s.ID, &s.UserID, &exp, &created, &rev)
	if e != nil {
		return s, domain.ErrUnauthorized
	}
	s.ExpiresAt, _ = time.Parse(time.RFC3339Nano, exp)
	s.CreatedAt, _ = time.Parse(time.RFC3339Nano, created)
	if rev.Valid || !s.ExpiresAt.After(now) {
		return s, domain.ErrUnauthorized
	}
	return s, nil
}
func (r Sessions) Revoke(ctx context.Context, id string, at time.Time) error {
	_, e := r.DB.ExecContext(ctx, "UPDATE sessions SET revoked_at=? WHERE id=?", at.Format(time.RFC3339Nano), id)
	return e
}
func (r Sessions) RevokeUser(ctx context.Context, userID string, at time.Time) error {
	_, e := r.DB.ExecContext(ctx, "UPDATE sessions SET revoked_at=? WHERE user_id=? AND revoked_at IS NULL", at.Format(time.RFC3339Nano), userID)
	return e
}
