package repository

import (
	"context"
	"github.com/11DingKing/dual-teacher-practice-go/internal/domain"
	"github.com/11DingKing/dual-teacher-practice-go/internal/storage/sqlite"
	"testing"
	"time"
)

func userSessionDB(t *testing.T) (Users, Sessions, *sqlite.DB) {
	db, e := sqlite.Open(context.Background(), ":memory:")
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { db.Close() })
	return Users{DB: db.SQL}, Sessions{DB: db.SQL}, db
}
func TestUserCreateFindAndRevoke(t *testing.T) {
	u, _, db := userSessionDB(t)
	now := time.Now().UTC()
	model := domain.User{ID: "u", Username: "teacher", DisplayName: "教师", Role: domain.RoleTeacher, CreatedAt: now}
	if e := u.Create(context.Background(), model, "hash"); e != nil {
		t.Fatal(e)
	}
	got, hash, e := u.FindByUsername(context.Background(), "teacher")
	if e != nil || hash != "hash" || got.Role != domain.RoleTeacher {
		t.Fatalf("%+v %s %v", got, hash, e)
	}
	if e = u.Revoke(context.Background(), "u", now); e != nil {
		t.Fatal(e)
	}
	got, e = u.Find(context.Background(), "u")
	if e != nil || got.RevokedAt == nil {
		t.Fatalf("%+v %v", got, e)
	}
	var revoked string
	if e = db.SQL.QueryRow("SELECT revoked_at FROM users WHERE id='u'").Scan(&revoked); e != nil || revoked == "" {
		t.Fatalf("%s %v", revoked, e)
	}
}
func TestUserFindMissing(t *testing.T) {
	u, _, _ := userSessionDB(t)
	if _, e := u.Find(context.Background(), "missing"); e == nil {
		t.Fatal("missing user found")
	}
	if _, _, e := u.FindByUsername(context.Background(), "missing"); e != domain.ErrNotFound {
		t.Fatalf("%v", e)
	}
}
func TestSessionLifecycle(t *testing.T) {
	u, s, _ := userSessionDB(t)
	now := time.Now().UTC()
	if e := u.Create(context.Background(), domain.User{ID: "u", Username: "u", DisplayName: "U", Role: domain.RoleTeacher, CreatedAt: now}, "p"); e != nil {
		t.Fatal(e)
	}
	session := domain.Session{ID: "s", UserID: "u", CreatedAt: now, ExpiresAt: now.Add(time.Hour)}
	if e := s.Create(context.Background(), session); e != nil {
		t.Fatal(e)
	}
	got, e := s.GetActive(context.Background(), "s", now)
	if e != nil || got.UserID != "u" {
		t.Fatalf("%+v %v", got, e)
	}
	if e = s.Revoke(context.Background(), "s", now); e != nil {
		t.Fatal(e)
	}
	if _, e = s.GetActive(context.Background(), "s", now); e != domain.ErrUnauthorized {
		t.Fatalf("%v", e)
	}
}
func TestSessionExpiryAndUserRevoke(t *testing.T) {
	u, s, _ := userSessionDB(t)
	now := time.Now().UTC()
	if e := u.Create(context.Background(), domain.User{ID: "u", Username: "u", DisplayName: "U", Role: domain.RoleTeacher, CreatedAt: now}, "p"); e != nil {
		t.Fatal(e)
	}
	for i := 0; i < 3; i++ {
		id := string(rune('a' + i))
		if e := s.Create(context.Background(), domain.Session{ID: id, UserID: "u", CreatedAt: now, ExpiresAt: now.Add(time.Hour)}); e != nil {
			t.Fatal(e)
		}
	}
	if _, e := s.GetActive(context.Background(), "missing", now); e != domain.ErrUnauthorized {
		t.Fatalf("%v", e)
	}
	expired := domain.Session{ID: "expired", UserID: "u", CreatedAt: now.Add(-2 * time.Hour), ExpiresAt: now.Add(-time.Hour)}
	if e := s.Create(context.Background(), expired); e != nil {
		t.Fatal(e)
	}
	if _, e := s.GetActive(context.Background(), "expired", now); e != domain.ErrUnauthorized {
		t.Fatalf("%v", e)
	}
	if e := s.RevokeUser(context.Background(), "u", now); e != nil {
		t.Fatal(e)
	}
	for _, id := range []string{"a", "b", "c"} {
		if _, e := s.GetActive(context.Background(), id, now); e != domain.ErrUnauthorized {
			t.Fatalf("%s %v", id, e)
		}
	}
}
func TestUserSeedIsIdempotent(t *testing.T) {
	u, _, _ := userSessionDB(t)
	now := time.Now().UTC()
	model := domain.User{ID: "u", Username: "u", DisplayName: "U", Role: domain.RoleTeacher, CreatedAt: now}
	if e := u.Seed(context.Background(), model, "p"); e != nil {
		t.Fatal(e)
	}
	if e := u.Seed(context.Background(), model, "p"); e != nil {
		t.Fatal(e)
	}
}
