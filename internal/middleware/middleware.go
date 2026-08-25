package middleware

import (
	"context"
	"fmt"
	"github.com/11DingKing/dual-teacher-practice-go/internal/domain"
	"github.com/11DingKing/dual-teacher-practice-go/internal/service"
	"net/http"
	"time"
)

type key string

const (
	UserKey      key = "user"
	RequestIDKey key = "request_id"
)

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = "req-" + serviceID()
		}
		ctx := context.WithValue(r.Context(), RequestIDKey, id)
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
func serviceID() string { return fmt.Sprintf("%d", time.Now().UnixNano()) }
func RequireAuth(auth service.Auth, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sid := r.Header.Get("Authorization")
		if len(sid) > 7 && sid[:7] == "Bearer " {
			sid = sid[7:]
		}
		u, e := auth.Authenticate(r.Context(), sid)
		if e != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), UserKey, u)))
	})
}
func RequireRole(roles ...domain.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, ok := r.Context().Value(UserKey).(domain.User)
			if !ok {
				http.Error(w, "unauthorized", 401)
				return
			}
			for _, role := range roles {
				if u.Role == role {
					next.ServeHTTP(w, r)
					return
				}
			}
			http.Error(w, "forbidden", 403)
		})
	}
}
