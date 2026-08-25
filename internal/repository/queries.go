package repository

import (
	"context"
	"database/sql"
	"github.com/11DingKing/dual-teacher-practice-go/internal/domain"
	"time"
)

type Queries struct{ DB *sql.DB }

func (r Queries) DueApplications(ctx context.Context, now time.Time, limit int) ([]domain.Application, error) {
	rows, e := r.DB.QueryContext(ctx, "SELECT id,teacher_id,quota_id,year,status,requested_hours,requested_budget_cents,idempotency_key,version,due_at,created_at,updated_at FROM applications WHERE due_at<=? AND status NOT IN ('certified','rejected','expired') ORDER BY due_at LIMIT ?", now.Format(time.RFC3339Nano), limit)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var out []domain.Application
	for rows.Next() {
		var a domain.Application
		var st, d, c, u string
		if e := rows.Scan(&a.ID, &a.TeacherID, &a.QuotaID, &a.Year, &st, &a.RequestedHours, &a.RequestedBudgetCents, &a.IdempotencyKey, &a.Version, &d, &c, &u); e != nil {
			return nil, e
		}
		a.Status = domain.ApplicationStatus(st)
		a.DueAt, _ = time.Parse(time.RFC3339Nano, d)
		a.CreatedAt, _ = time.Parse(time.RFC3339Nano, c)
		a.UpdatedAt, _ = time.Parse(time.RFC3339Nano, u)
		out = append(out, a)
	}
	return out, rows.Err()
}
func (r Queries) EvidenceSummary(ctx context.Context, id string) (map[string]int, error) {
	rows, e := r.DB.QueryContext(ctx, "SELECT verification_status,COUNT(*) FROM evidences WHERE application_id=? GROUP BY verification_status", id)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := map[string]int{}
	for rows.Next() {
		var k string
		var n int
		if e := rows.Scan(&k, &n); e != nil {
			return nil, e
		}
		out[k] = n
	}
	return out, rows.Err()
}
