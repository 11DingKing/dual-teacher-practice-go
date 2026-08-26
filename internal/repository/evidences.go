package repository

import (
	"context"
	"database/sql"
	"github.com/11DingKing/dual-teacher-practice-go/internal/domain"
	"time"
)

type Evidences struct{ DB *sql.DB }

func (r Evidences) Add(ctx context.Context, e domain.Evidence) error {
	_, x := r.DB.ExecContext(ctx, "INSERT INTO evidences(id,application_id,kind,title,uri,verification_status,submitted_by,created_at) VALUES(?,?,?,?,?,?,?,?)", e.ID, e.ApplicationID, e.Kind, e.Title, e.URI, e.Status, e.SubmittedBy, e.CreatedAt.Format(time.RFC3339Nano))
	return x
}
func (r Evidences) SetStatus(ctx context.Context, id string, from, to domain.EvidenceStatus, actor string, at time.Time, note string) error {
	res, e := r.DB.ExecContext(ctx, "UPDATE evidences SET verification_status=?,verified_by=?,verified_at=?,dispute_note=? WHERE id=? AND verification_status=?", to, actor, at.Format(time.RFC3339Nano), note, id, from)
	if e != nil {
		return e
	}
	n, _ := res.RowsAffected()
	if n != 1 {
		return domain.ErrConflict
	}
	return nil
}
func (r Evidences) ForApplication(ctx context.Context, id string) ([]domain.Evidence, error) {
	rows, e := r.DB.QueryContext(ctx, "SELECT id,application_id,kind,title,uri,verification_status,submitted_by,created_at FROM evidences WHERE application_id=? ORDER BY created_at", id)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var out []domain.Evidence
	for rows.Next() {
		var v domain.Evidence
		var st, created string
		if e := rows.Scan(&v.ID, &v.ApplicationID, &v.Kind, &v.Title, &v.URI, &st, &v.SubmittedBy, &created); e != nil {
			return nil, e
		}
		v.Status = domain.EvidenceStatus(st)
		v.CreatedAt, _ = time.Parse(time.RFC3339Nano, created)
		out = append(out, v)
	}
	return out, rows.Err()
}
