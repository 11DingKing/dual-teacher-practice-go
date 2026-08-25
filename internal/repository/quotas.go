package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/11DingKing/dual-teacher-practice-go/internal/domain"
)

type Quotas struct{ DB *sql.DB }

func (r Quotas) GetForUpdate(ctx context.Context, tx *sql.Tx, id string) (domain.Quota, error) {
	var q domain.Quota
	var active int
	e := tx.QueryRowContext(ctx, "SELECT id,name,max_hours,used_hours,max_budget_cents,used_budget_cents,version,active FROM quotas WHERE id=?", id).Scan(&q.ID, &q.Name, &q.MaxHours, &q.UsedHours, &q.MaxBudgetCents, &q.UsedBudgetCents, &q.Version, &active)
	if e != nil {
		return q, fmt.Errorf("quota: %w", e)
	}
	q.Active = active == 1
	return q, nil
}
func (r Quotas) Create(ctx context.Context, q domain.Quota) error {
	a := 0
	if q.Active {
		a = 1
	}
	_, e := r.DB.ExecContext(ctx, "INSERT INTO quotas(id,name,max_hours,used_hours,max_budget_cents,used_budget_cents,version,active) VALUES(?,?,?,?,?,?,?,?)", q.ID, q.Name, q.MaxHours, q.UsedHours, q.MaxBudgetCents, q.UsedBudgetCents, q.Version, a)
	return e
}
func (r Quotas) Release(ctx context.Context, tx *sql.Tx, id string, h, b, version int) error {
	res, e := tx.ExecContext(ctx, "UPDATE quotas SET used_hours=used_hours-?,used_budget_cents=used_budget_cents-?,version=version+1 WHERE id=? AND version=? AND used_hours>=? AND used_budget_cents>=?", h, b, id, version, h, b)
	if e != nil {
		return e
	}
	n, _ := res.RowsAffected()
	if n != 1 {
		return domain.ErrConflict
	}
	return nil
}
