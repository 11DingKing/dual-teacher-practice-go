package service

import (
	"context"
	"github.com/11DingKing/dual-teacher-practice-go/internal/clock"
	"github.com/11DingKing/dual-teacher-practice-go/internal/domain"
	"github.com/11DingKing/dual-teacher-practice-go/internal/repository"
	"github.com/11DingKing/dual-teacher-practice-go/internal/storage/sqlite"
	"testing"
	"time"
)

func authFixture(t *testing.T) (Auth, *sqlite.DB, domain.User) {
	t.Helper()
	db, e := sqlite.Open(context.Background(), ":memory:")
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { db.Close() })
	now := time.Date(2026, 8, 25, 0, 0, 0, 0, time.UTC)
	u := domain.User{ID: "u1", Username: "teacher", DisplayName: "教师", Role: domain.RoleTeacher, CreatedAt: now}
	users := repository.Users{DB: db.SQL}
	if e = users.Create(context.Background(), u, hashPassword("pw")); e != nil {
		t.Fatal(e)
	}
	return Auth{Users: users, Sessions: repository.Sessions{DB: db.SQL}, Clock: clock.Fixed{T: now}, TTL: time.Hour}, db, u
}
func TestLoginAndAuthenticate(t *testing.T) {
	a, _, u := authFixture(t)
	s, got, e := a.Login(context.Background(), "teacher", "pw")
	if e != nil {
		t.Fatal(e)
	}
	if got.ID != u.ID || s.ID == "" {
		t.Fatalf("%+v %+v", got, s)
	}
	again, e := a.Authenticate(context.Background(), s.ID)
	if e != nil || again.ID != u.ID {
		t.Fatalf("%+v %v", again, e)
	}
}
func TestLoginRejectsBadPassword(t *testing.T) {
	a, _, _ := authFixture(t)
	if _, _, e := a.Login(context.Background(), "teacher", "wrong"); e != domain.ErrUnauthorized {
		t.Fatalf("%v", e)
	}
}
func TestLogoutRevokesSession(t *testing.T) {
	a, _, _ := authFixture(t)
	s, _, e := a.Login(context.Background(), "teacher", "pw")
	if e != nil {
		t.Fatal(e)
	}
	if e = a.Logout(context.Background(), s.ID); e != nil {
		t.Fatal(e)
	}
	if _, e = a.Authenticate(context.Background(), s.ID); e != domain.ErrUnauthorized {
		t.Fatalf("%v", e)
	}
}
func TestExpiredSessionRejected(t *testing.T) {
	a, db, u := authFixture(t)
	old := domain.Session{ID: "old", UserID: u.ID, CreatedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), ExpiresAt: time.Date(2025, 1, 1, 1, 0, 0, 0, time.UTC)}
	if e := (repository.Sessions{DB: db.SQL}).Create(context.Background(), old); e != nil {
		t.Fatal(e)
	}
	if _, e := a.Authenticate(context.Background(), old.ID); e != domain.ErrUnauthorized {
		t.Fatalf("%v", e)
	}
}
func TestRevokedUserCannotLogin(t *testing.T) {
	a, db, u := authFixture(t)
	if e := (repository.Users{DB: db.SQL}).Revoke(context.Background(), u.ID, time.Now().UTC()); e != nil {
		t.Fatal(e)
	}
	if _, _, e := a.Login(context.Background(), u.Username, "pw"); e != domain.ErrUnauthorized {
		t.Fatalf("%v", e)
	}
}
func TestSessionUserIsolation(t *testing.T) {
	a, db, u := authFixture(t)
	other := domain.User{ID: "u2", Username: "hr", DisplayName: "人事", Role: domain.RoleHR, CreatedAt: u.CreatedAt}
	users := repository.Users{DB: db.SQL}
	if e := users.Create(context.Background(), other, hashPassword("hr")); e != nil {
		t.Fatal(e)
	}
	s, _, e := a.Login(context.Background(), "hr", "hr")
	if e != nil {
		t.Fatal(e)
	}
	got, e := a.Authenticate(context.Background(), s.ID)
	if e != nil || got.ID != other.ID {
		t.Fatalf("%+v %v", got, e)
	}
}
