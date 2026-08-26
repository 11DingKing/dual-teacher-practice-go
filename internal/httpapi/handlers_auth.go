package httpapi

import (
	"encoding/json"
	"net/http"
)

func (s Server) authRoutes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/auth/login", s.login)
	mux.HandleFunc("/api/auth/logout", s.logout)
	return mux
}
func (s Server) login(w http.ResponseWriter, r *http.Request) {
	var in struct{ Username, Password string }
	if json.NewDecoder(r.Body).Decode(&in) != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid json"})
		return
	}
	sess, user, e := s.Auth.Login(r.Context(), in.Username, in.Password)
	if e != nil {
		writeJSON(w, 401, map[string]string{"error": "unauthorized"})
		return
	}
	writeJSON(w, 200, map[string]any{"token": sess.ID, "expires_at": sess.ExpiresAt, "user": user})
}
func (s Server) logout(w http.ResponseWriter, r *http.Request) {
	sid := r.Header.Get("Authorization")
	if len(sid) > 7 {
		sid = sid[7:]
	}
	if e := s.Auth.Logout(r.Context(), sid); e != nil {
		writeJSON(w, 400, map[string]string{"error": e.Error()})
		return
	}
	writeJSON(w, 200, map[string]string{"status": "revoked"})
}
