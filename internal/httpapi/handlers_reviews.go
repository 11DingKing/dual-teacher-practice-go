package httpapi

import (
	"encoding/json"
	"github.com/11DingKing/dual-teacher-practice-go/internal/domain"
	"github.com/11DingKing/dual-teacher-practice-go/internal/middleware"
	"net/http"
	"strings"
)

func (s Server) review(w http.ResponseWriter, r *http.Request) {
	u := r.Context().Value(middleware.UserKey).(domain.User)
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 4 {
		writeJSON(w, 404, nil)
		return
	}
	id := parts[3]
	var in struct {
		Score    int
		Decision domain.ReviewDecision
		Comment  string
	}
	if json.NewDecoder(r.Body).Decode(&in) != nil {
		writeJSON(w, 400, nil)
		return
	}
	if e := s.Review.Add(r.Context(), domain.Review{ID: id, ApplicationID: in.Comment, Score: in.Score, Decision: in.Decision, Comment: in.Comment, ReviewerRole: u.Role}, u.ID, r.Context().Value(middleware.RequestIDKey).(string)); e != nil {
		writeJSON(w, 409, map[string]string{"error": e.Error()})
		return
	}
	writeJSON(w, 201, map[string]string{"status": "recorded"})
}
