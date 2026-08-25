package service

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/11DingKing/dual-teacher-practice-go/internal/clock"
	"github.com/11DingKing/dual-teacher-practice-go/internal/domain"
	"github.com/11DingKing/dual-teacher-practice-go/internal/repository"
	"time"
)

type Application struct {
	DB     *sql.DB
	Apps   repository.Applications
	Quotas repository.Quotas
	Audit  repository.Audit
	Clock  clock.Clock
}

func (s Application) Submit(ctx context.Context, a domain.Application, actor, requestID string) (domain.Application, error) {
	if a.RequestedHours <= 0 || a.RequestedBudgetCents < 0 {
		return a, fmt.Errorf("%w: invalid request", domain.ErrConflict)
	}
	now := s.Clock.Now()
	a.Status = domain.StatusDraft
	a.Version = 1
	a.CreatedAt = now
	a.UpdatedAt = now
	a.DueAt = now.AddDate(0, 3, 0)
	tx, e := s.DB.BeginTx(ctx, nil)
	if e != nil {
		return a, e
	}
	defer tx.Rollback()
	q, e := s.Quotas.GetForUpdate(ctx, tx, a.QuotaID)
	if e != nil {
		return a, e
	}
	if !q.Active || q.UsedHours+a.RequestedHours > q.MaxHours || q.UsedBudgetCents+a.RequestedBudgetCents > q.MaxBudgetCents {
		return a, domain.ErrQuotaExceeded
	}
	if _, e = tx.ExecContext(ctx, "UPDATE quotas SET used_hours=used_hours+?,used_budget_cents=used_budget_cents+?,version=version+1 WHERE id=? AND version=?", a.RequestedHours, a.RequestedBudgetCents, q.ID, q.Version); e != nil {
		return a, e
	}
	if e = s.Apps.CreateTx(ctx, tx, a); e != nil {
		return a, e
	}
	ev := domain.AuditEvent{ID: token(), ActorID: actor, Action: "application_created", EntityType: "application", EntityID: a.ID, Outcome: "success", RequestID: requestID, Details: "quota reserved", CreatedAt: now}
	if e = s.auditTx(ctx, tx, ev); e != nil {
		return a, e
	}
	if e = tx.Commit(); e != nil {
		return a, e
	}
	return a, nil
}
func (s Application) auditTx(ctx context.Context, tx *sql.Tx, e domain.AuditEvent) error {
	_, err := tx.ExecContext(ctx, "INSERT INTO audit_events(id,actor_id,action,entity_type,entity_id,outcome,request_id,details,created_at) VALUES(?,?,?,?,?,?,?,?,?)", e.ID, e.ActorID, e.Action, e.EntityType, e.EntityID, e.Outcome, e.RequestID, e.Details, e.CreatedAt.Format(time.RFC3339Nano))
	return err
}
func (s Application) Transition(ctx context.Context, id string, to domain.ApplicationStatus, actor, requestID string) error {
	a, e := s.Apps.Get(ctx, id)
	if e != nil {
		return e
	}
	if s.Clock.Now().After(a.DueAt) && to != domain.StatusRenewal {
		return domain.ErrDeadline
	}
	if e = s.Apps.Transition(ctx, id, a.Status, to, s.Clock.Now()); e != nil {
		return e
	}
	return s.Audit.Append(ctx, domain.AuditEvent{ID: token(), ActorID: actor, Action: "application_transition", EntityType: "application", EntityID: id, Outcome: "success", RequestID: requestID, Details: fmt.Sprintf("%s -> %s", a.Status, to), CreatedAt: s.Clock.Now()})
}
func (s Application) Get(ctx context.Context, id string) (domain.Application, error) {
	return s.Apps.Get(ctx, id)
}
