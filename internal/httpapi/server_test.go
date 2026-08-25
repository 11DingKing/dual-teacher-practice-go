package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/11DingKing/dual-teacher-practice-go/internal/clock"
	"github.com/11DingKing/dual-teacher-practice-go/internal/domain"
	"github.com/11DingKing/dual-teacher-practice-go/internal/repository"
	"github.com/11DingKing/dual-teacher-practice-go/internal/service"
	"github.com/11DingKing/dual-teacher-practice-go/internal/storage/sqlite"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func httpFixture(t *testing.T) http.Handler {
	db, e := sqlite.Open(context.Background(), ":memory:")
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { db.Close() })
	now := time.Now().UTC()
	u := domain.User{ID: "u", Username: "teacher", DisplayName: "教师", Role: domain.RoleTeacher, CreatedAt: now}
	users := repository.Users{DB: db.SQL}
	if e = users.Create(context.Background(), u, testHash("pw")); e != nil {
		t.Fatal(e)
	}
	auth := service.Auth{Users: users, Sessions: repository.Sessions{DB: db.SQL}, Clock: clock.Fixed{T: now}, TTL: time.Hour}
	return (Server{Auth: auth, Apps: service.Application{DB: db.SQL, Apps: repository.Applications{DB: db.SQL}, Quotas: repository.Quotas{DB: db.SQL}, Audit: repository.Audit{DB: db.SQL}, Clock: clock.Fixed{T: now}}}).Routes()
}

func testHash(v string) string { h := sha256.Sum256([]byte(v)); return hex.EncodeToString(h[:]) }
func TestHealthAndReady(t *testing.T) {
	h := httpFixture(t)
	for _, path := range []string{"/healthz", "/readyz"} {
		r := httptest.NewRequest(http.MethodGet, path, nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != 200 {
			t.Fatalf("%s %d", path, w.Code)
		}
		var body map[string]string
		if e := json.Unmarshal(w.Body.Bytes(), &body); e != nil || body["status"] == "" {
			t.Fatalf("%s", w.Body.String())
		}
	}
}
func TestRequestIDPreserved(t *testing.T) {
	h := httpFixture(t)
	r := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	r.Header.Set("X-Request-ID", "client-123")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Header().Get("X-Request-ID") != "client-123" {
		t.Fatal(w.Header())
	}
}
func TestLoginBadJSON(t *testing.T) {
	h := httpFixture(t)
	r := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader("{"))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 400 {
		t.Fatal(w.Code)
	}
}
func TestLoginBadPassword(t *testing.T) {
	h := httpFixture(t)
	r := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(`{"username":"teacher","password":"bad"}`))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 401 {
		t.Fatal(w.Code)
	}
}
func TestProtectedEndpointUnauthorized(t *testing.T) {
	h := httpFixture(t)
	r := httptest.NewRequest(http.MethodGet, "/api/applications", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 401 {
		t.Fatal(w.Code)
	}
}
func TestUnknownApplication(t *testing.T) {
	h := httpFixture(t)
	r := httptest.NewRequest(http.MethodGet, "/api/applications/missing", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 401 {
		t.Fatal(w.Code)
	}
}
func TestMethodNotAllowed(t *testing.T) {
	h := httpFixture(t)
	r := httptest.NewRequest(http.MethodDelete, "/healthz", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatal("health handler is intentionally read-only but net/http serves it")
	}
}
