package httpapi

import (
	"encoding/json"
	"github.com/11DingKing/dual-teacher-practice-go/internal/domain"
	"github.com/11DingKing/dual-teacher-practice-go/internal/middleware"
	"net/http"
	"strings"
	"time"
)

func (s Server) practice(w http.ResponseWriter, r *http.Request) {
	u := r.Context().Value(middleware.UserKey).(domain.User)
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 4 {
		writeError(w, domain.ErrNotFound)
		return
	}
	id := parts[3]
	if r.Method == "POST" {
		var in struct {
			ApplicationID, CompanyName, MentorName, Email string
			Hours                                         int
			StartsAt, EndsAt                              string
		}
		if e := json.NewDecoder(r.Body).Decode(&in); e != nil {
			writeError(w, e)
			return
		}
		starts, _ := time.Parse(time.RFC3339, in.StartsAt)
		ends, _ := time.Parse(time.RFC3339, in.EndsAt)
		p := domain.Practice{ID: id, ApplicationID: in.ApplicationID, CompanyName: in.CompanyName, MentorName: in.MentorName, ContactEmail: in.Email, PlannedHours: in.Hours, StartsAt: starts, EndsAt: ends}
		if e := s.Practice.Register(r.Context(), p, u.ID, r.Context().Value(middleware.RequestIDKey).(string)); e != nil {
			writeError(w, e)
			return
		}
		writeJSON(w, 201, p)
		return
	}
	writeError(w, domain.ErrConflict)
}
