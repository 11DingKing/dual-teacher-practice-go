package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/11DingKing/dual-teacher-practice-go/internal/domain"
	"time"
)

type Applications struct{ DB *sql.DB }

func (r Applications) CreateTx(ctx context.Context, tx *sql.Tx, a domain.Application) error {
	_, e := tx.ExecContext(ctx, "INSERT INTO applications(id,teacher_id,quota_id,year,status,requested_hours,requested_budget_cents,idempotency_key,version,due_at,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?)", a.ID, a.TeacherID, a.QuotaID, a.Year, a.Status, a.RequestedHours, a.RequestedBudgetCents, a.IdempotencyKey, a.Version, a.DueAt.Format(time.RFC3339Nano), a.CreatedAt.Format(time.RFC3339Nano), a.UpdatedAt.Format(time.RFC3339Nano))
	return e
}
func (r Applications) Get(ctx context.Context, id string) (domain.Application, error) {
	var a domain.Application
	var status, due, created, updated string
	e := r.DB.QueryRowContext(ctx, "SELECT id,teacher_id,quota_id,year,status,requested_hours,requested_budget_cents,idempotency_key,version,due_at,created_at,updated_at FROM applications WHERE id=?", id).Scan(&a.ID, &a.TeacherID, &a.QuotaID, &a.Year, &status, &a.RequestedHours, &a.RequestedBudgetCents, &a.IdempotencyKey, &a.Version, &due, &created, &updated)
	if e != nil {
		if e == sql.ErrNoRows {
			return a, domain.ErrNotFound
		}
		return a, e
	}
	a.Status = domain.ApplicationStatus(status)
	a.DueAt, _ = time.Parse(time.RFC3339Nano, due)
	a.CreatedAt, _ = time.Parse(time.RFC3339Nano, created)
	a.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updated)
	return a, nil
}
func (r Applications) Transition(ctx context.Context, id string, from, to domain.ApplicationStatus, now time.Time) error {
	if err := domain.Transition(from, to); err != nil {
		return err
	}
	res, e := r.DB.ExecContext(ctx, "UPDATE applications SET status=?,version=version+1,updated_at=? WHERE id=? AND status=?", to, now.Format(time.RFC3339Nano), id, from)
	if e != nil {
		return fmt.Errorf("transition: %w", e)
	}
	n, _ := res.RowsAffected()
	if n != 1 {
		return domain.ErrConflict
	}
	return nil
}
func (r Applications) List(ctx context.Context, status string, limit, offset int) ([]domain.Application, error) {
	rows, e := r.DB.QueryContext(ctx, "SELECT id,teacher_id,quota_id,year,status,requested_hours,requested_budget_cents,idempotency_key,version,due_at,created_at,updated_at FROM applications WHERE (?='' OR status=?) ORDER BY created_at DESC LIMIT ? OFFSET ?", status, status, limit, offset)
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
