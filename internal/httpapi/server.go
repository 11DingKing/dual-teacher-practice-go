package httpapi

import (
	"encoding/json"
	"github.com/11DingKing/dual-teacher-practice-go/internal/middleware"
	"github.com/11DingKing/dual-teacher-practice-go/internal/service"
	"net/http"
)

type Server struct {
	Auth     service.Auth
	Apps     service.Application
	Evidence service.Evidence
	Practice service.Practice
	Review   service.Review
	Report   service.Report
}

func (s Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
	})
	mux.Handle("/api/auth/", s.authRoutes())
	mux.Handle("/api/applications", middleware.RequireAuth(s.Auth, http.HandlerFunc(s.applications)))
	mux.Handle("/api/applications/", middleware.RequireAuth(s.Auth, http.HandlerFunc(s.applicationDetail)))
	return middleware.RequestID(recoverer(mux))
}
func recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recover() != nil {
				http.Error(w, "internal error", 500)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
