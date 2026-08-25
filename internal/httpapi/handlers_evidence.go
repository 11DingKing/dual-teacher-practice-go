package httpapi

import (
	"encoding/json"
	"github.com/11DingKing/dual-teacher-practice-go/internal/domain"
	"github.com/11DingKing/dual-teacher-practice-go/internal/middleware"
	"net/http"
	"strings"
)

func (s Server) evidence(w http.ResponseWriter, r *http.Request) {
	u := r.Context().Value(middleware.UserKey).(domain.User)
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 4 {
		writeJSON(w, 404, nil)
		return
	}
	id := parts[3]
	if r.Method == "POST" {
		var in struct{ ApplicationID, Kind, Title, URI string }
		if json.NewDecoder(r.Body).Decode(&in) != nil {
			writeJSON(w, 400, nil)
			return
		}
		e := domain.Evidence{ID: id, ApplicationID: in.ApplicationID, Kind: in.Kind, Title: in.Title, URI: in.URI}
		if err := s.Evidence.Add(r.Context(), e, u.ID, r.Context().Value(middleware.RequestIDKey).(string)); err != nil {
			writeJSON(w, 409, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, 201, e)
		return
	}
	writeJSON(w, 405, nil)
}
