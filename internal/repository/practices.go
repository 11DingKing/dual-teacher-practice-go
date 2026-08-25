package repository

import (
	"context"
	"database/sql"
	"github.com/11DingKing/dual-teacher-practice-go/internal/domain"
	"time"
)

type Practices struct{ DB *sql.DB }

// BeginTx starts a transaction bound to ctx so a cancelled or timed-out
// request aborts every statement executed within it before any commit.
func (r Practices) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return r.DB.BeginTx(ctx, nil)
}

func (r Practices) Add(ctx context.Context, p domain.Practice) error {
	_, e := r.DB.ExecContext(ctx, "INSERT INTO practices(id,application_id,company_name,mentor_name,contact_email,status,planned_hours,actual_hours,starts_at,ends_at,version) VALUES(?,?,?,?,?,?,?,?,?,?,?)", p.ID, p.ApplicationID, p.CompanyName, p.MentorName, p.ContactEmail, p.Status, p.PlannedHours, p.ActualHours, p.StartsAt.Format(time.RFC3339Nano), p.EndsAt.Format(time.RFC3339Nano), p.Version)
	return e
}
func (r Practices) Get(ctx context.Context, id string) (domain.Practice, error) {
	var p domain.Practice
	var st, s, e string
	err := r.DB.QueryRowContext(ctx, "SELECT id,application_id,company_name,mentor_name,contact_email,status,planned_hours,actual_hours,starts_at,ends_at,version FROM practices WHERE id=?", id).Scan(&p.ID, &p.ApplicationID, &p.CompanyName, &p.MentorName, &p.ContactEmail, &st, &p.PlannedHours, &p.ActualHours, &s, &e, &p.Version)
	if err != nil {
		return p, err
	}
	p.Status = st
	p.StartsAt, _ = time.Parse(time.RFC3339Nano, s)
	p.EndsAt, _ = time.Parse(time.RFC3339Nano, e)
	return p, nil
}
func (r Practices) UpdateHours(ctx context.Context, id string, h, version int) error {
	res, e := r.DB.ExecContext(ctx, "UPDATE practices SET actual_hours=?,version=version+1 WHERE id=? AND version=? AND actual_hours<=planned_hours", h, id, version)
	if e != nil {
		return e
	}
	n, _ := res.RowsAffected()
	if n != 1 {
		return domain.ErrConflict
	}
	return nil
}

// UpdateHoursTx applies the hours update within the caller's transaction so the
// change commits or rolls back together with the audit event. It honours ctx:
// a cancelled or timed-out request aborts the statement before any row is
// mutated.
func (r Practices) UpdateHoursTx(ctx context.Context, tx *sql.Tx, id string, h, version int) error {
	res, e := tx.ExecContext(ctx, "UPDATE practices SET actual_hours=?,version=version+1 WHERE id=? AND version=? AND actual_hours<=planned_hours", h, id, version)
	if e != nil {
		return e
	}
	n, _ := res.RowsAffected()
	if n != 1 {
		return domain.ErrConflict
	}
	return nil
}
