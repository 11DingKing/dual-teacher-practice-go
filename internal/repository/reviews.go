package repository

import (
	"context"
	"database/sql"
	"github.com/11DingKing/dual-teacher-practice-go/internal/domain"
	"time"
)

type Reviews struct{ DB *sql.DB }

func (r Reviews) Add(ctx context.Context, v domain.Review) error {
	_, e := r.DB.ExecContext(ctx, "INSERT INTO reviews(id,application_id,reviewer_id,reviewer_role,score,decision,comment,created_at) VALUES(?,?,?,?,?,?,?,?)", v.ID, v.ApplicationID, v.ReviewerID, v.ReviewerRole, v.Score, v.Decision, v.Comment, v.CreatedAt.Format(time.RFC3339Nano))
	return e
}
func (r Reviews) ForApplication(ctx context.Context, id string) ([]domain.Review, error) {
	rows, e := r.DB.QueryContext(ctx, "SELECT id,application_id,reviewer_id,reviewer_role,score,decision,comment,created_at FROM reviews WHERE application_id=? ORDER BY created_at", id)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var out []domain.Review
	for rows.Next() {
		var v domain.Review
		var role, dec, created string
		if e := rows.Scan(&v.ID, &v.ApplicationID, &v.ReviewerID, &role, &v.Score, &dec, &v.Comment, &created); e != nil {
			return nil, e
		}
		v.ReviewerRole = domain.Role(role)
		v.Decision = domain.ReviewDecision(dec)
		v.CreatedAt, _ = time.Parse(time.RFC3339Nano, created)
		out = append(out, v)
	}
	return out, rows.Err()
}
