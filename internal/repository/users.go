package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/11DingKing/dual-teacher-practice-go/internal/domain"
	"time"
)

type Users struct{ DB *sql.DB }

func (r Users) Create(ctx context.Context, u domain.User, passwordHash string) error {
	_, err := r.DB.ExecContext(ctx, "INSERT INTO users(id,username,password_hash,role,display_name,created_at) VALUES(?,?,?,?,?,?)", u.ID, u.Username, passwordHash, u.Role, u.DisplayName, u.CreatedAt.Format(time.RFC3339Nano))
	return err
}
func (r Users) FindByUsername(ctx context.Context, username string) (domain.User, string, error) {
	var u domain.User
	var role, created string
	var revoked sql.NullString
	var hash string
	err := r.DB.QueryRowContext(ctx, "SELECT id,username,password_hash,role,display_name,created_at,revoked_at FROM users WHERE username=?", username).Scan(&u.ID, &u.Username, &hash, &role, &u.DisplayName, &created, &revoked)
	if err != nil {
		if err == sql.ErrNoRows {
			return u, "", domain.ErrNotFound
		}
		return u, "", fmt.Errorf("find user: %w", err)
	}
	u.Role = domain.Role(role)
	u.CreatedAt, _ = time.Parse(time.RFC3339Nano, created)
	if revoked.Valid {
		t, _ := time.Parse(time.RFC3339Nano, revoked.String)
		u.RevokedAt = &t
	}
	return u, hash, nil
}
func (r Users) Find(ctx context.Context, id string) (domain.User, error) {
	var u domain.User
	var role, created string
	var revoked sql.NullString
	err := r.DB.QueryRowContext(ctx, "SELECT id,username,role,display_name,created_at,revoked_at FROM users WHERE id=?", id).Scan(&u.ID, &u.Username, &role, &u.DisplayName, &created, &revoked)
	if err != nil {
		return u, err
	}
	u.Role = domain.Role(role)
	u.CreatedAt, _ = time.Parse(time.RFC3339Nano, created)
	if revoked.Valid {
		t, _ := time.Parse(time.RFC3339Nano, revoked.String)
		u.RevokedAt = &t
	}
	return u, nil
}
func (r Users) Revoke(ctx context.Context, id string, at time.Time) error {
	_, err := r.DB.ExecContext(ctx, "UPDATE users SET revoked_at=? WHERE id=?", at.Format(time.RFC3339Nano), id)
	return err
}
func (r Users) Seed(ctx context.Context, u domain.User, passwordHash string) error {
	if _, err := r.Find(ctx, u.ID); err == nil {
		return nil
	}
	return r.Create(ctx, u, passwordHash)
}
