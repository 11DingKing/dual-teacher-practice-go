package httpapi

import (
	"encoding/json"
	"github.com/11DingKing/dual-teacher-practice-go/internal/domain"
	"github.com/11DingKing/dual-teacher-practice-go/internal/middleware"
	"github.com/11DingKing/dual-teacher-practice-go/internal/pagination"
	"net/http"
	"strings"
	"time"
)

func (s Server) applications(w http.ResponseWriter, r *http.Request) {
	u := r.Context().Value(middleware.UserKey).(domain.User)
	if r.Method == "GET" {
		items, e := s.Report.List(r.Context(), r.URL.Query().Get("status"), pagination.Parse(r.URL.Query().Get("limit"), r.URL.Query().Get("offset")))
		if e != nil {
			writeJSON(w, 500, map[string]string{"error": e.Error()})
			return
		}
		writeJSON(w, 200, items)
		return
	}
	if r.Method != "POST" {
		writeJSON(w, 405, nil)
		return
	}
	var in struct {
		ID, QuotaID         string
		Year, Hours, Budget int
	}
	if json.NewDecoder(r.Body).Decode(&in) != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid json"})
		return
	}
	a := domain.Application{ID: in.ID, TeacherID: u.ID, QuotaID: in.QuotaID, Year: in.Year, RequestedHours: in.Hours, RequestedBudgetCents: in.Budget, IdempotencyKey: r.Header.Get("Idempotency-Key")}
	if a.IdempotencyKey == "" {
		a.IdempotencyKey = a.ID
	}
	v, e := s.Apps.Submit(r.Context(), a, u.ID, r.Context().Value(middleware.RequestIDKey).(string))
	if e != nil {
		writeJSON(w, 409, map[string]string{"error": e.Error()})
		return
	}
	writeJSON(w, 201, v)
}
func (s Server) applicationDetail(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 {
		writeJSON(w, 404, nil)
		return
	}
	id := parts[2]
	if len(parts) == 3 && r.Method == "GET" {
		a, e := s.Apps.Get(r.Context(), id)
		if e != nil {
			writeJSON(w, 404, map[string]string{"error": e.Error()})
			return
		}
		writeJSON(w, 200, a)
		return
	}
	if len(parts) == 4 && parts[3] == "transition" && r.Method == "POST" {
		var in struct{ To domain.ApplicationStatus }
		if json.NewDecoder(r.Body).Decode(&in) != nil {
			writeJSON(w, 400, nil)
			return
		}
		u := r.Context().Value(middleware.UserKey).(domain.User)
		if e := s.Apps.Transition(r.Context(), id, in.To, u.ID, r.Context().Value(middleware.RequestIDKey).(string)); e != nil {
			writeJSON(w, 409, map[string]string{"error": e.Error()})
			return
		}
		writeJSON(w, 200, map[string]string{"status": "updated", "at": time.Now().UTC().Format(time.RFC3339)})
		return
	}
	writeJSON(w, 404, nil)
}
