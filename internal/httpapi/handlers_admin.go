package httpapi

import (
	"encoding/json"
	"github.com/11DingKing/dual-teacher-practice-go/internal/domain"
	"github.com/11DingKing/dual-teacher-practice-go/internal/middleware"
	"github.com/11DingKing/dual-teacher-practice-go/internal/service"
	"net/http"
)

func AdminEndpoint(auth service.Auth, next func(http.ResponseWriter, *http.Request)) http.Handler {
	return middleware.RequireAuth(auth, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := r.Context().Value(middleware.UserKey).(domain.User)
		if u.Role != domain.RoleHR && u.Role != domain.RoleAdmin {
			writeError(w, domain.ErrForbidden)
			return
		}
		next(w, r)
	}))
}
func jsonBody(r *http.Request, v any) error { return json.NewDecoder(r.Body).Decode(v) }
