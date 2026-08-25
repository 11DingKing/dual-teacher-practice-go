package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/11DingKing/dual-teacher-practice-go/internal/clock"
	"github.com/11DingKing/dual-teacher-practice-go/internal/domain"
	"github.com/11DingKing/dual-teacher-practice-go/internal/repository"
	"time"
)

type Auth struct {
	Users    repository.Users
	Sessions repository.Sessions
	Clock    clock.Clock
	TTL      time.Duration
}

func hashPassword(v string) string { h := sha256.Sum256([]byte(v)); return hex.EncodeToString(h[:]) }
func token() string {
	b := make([]byte, 24)
	if _, e := rand.Read(b); e != nil {
		return fmt.Sprintf("fallback-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}
func (a Auth) Login(ctx context.Context, username, password string) (domain.Session, domain.User, error) {
	u, hash, e := a.Users.FindByUsername(ctx, username)
	if e != nil || hash != hashPassword(password) || u.RevokedAt != nil {
		return domain.Session{}, domain.User{}, domain.ErrUnauthorized
	}
	now := a.Clock.Now()
	s := domain.Session{ID: token(), UserID: u.ID, CreatedAt: now, ExpiresAt: now.Add(a.TTL)}
	if e = a.Sessions.Create(ctx, s); e != nil {
		return s, u, e
	}
	return s, u, nil
}
func (a Auth) Logout(ctx context.Context, sid string) error {
	return a.Sessions.Revoke(ctx, sid, a.Clock.Now())
}
func (a Auth) Authenticate(ctx context.Context, sid string) (domain.User, error) {
	s, e := a.Sessions.GetActive(ctx, sid, a.Clock.Now())
	if e != nil {
		return domain.User{}, e
	}
	return a.Users.Find(ctx, s.UserID)
}
