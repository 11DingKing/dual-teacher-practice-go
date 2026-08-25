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

type Practice struct {
	Practices repository.Practices
	Apps      repository.Applications
	Audit     repository.Audit
	Clock     clock.Clock
}

func (s Practice) Register(ctx context.Context, p domain.Practice, actor, requestID string) error {
	a, e := s.Apps.Get(ctx, p.ApplicationID)
	if e != nil {
		return e
	}
	if a.Status != domain.StatusCollegeApproved && a.Status != domain.StatusEvidenceReview {
		return domain.ErrInvalidState
	}
	if p.StartsAt.Before(s.Clock.Now().Add(-time.Minute)) || !p.EndsAt.After(p.StartsAt) {
		return fmt.Errorf("%w: practice window", domain.ErrConflict)
	}
	p.Status = "planned"
	p.Version = 1
	if e = s.Practices.Add(ctx, p); e != nil {
		return e
	}
	return s.Audit.Append(ctx, domain.AuditEvent{ID: token(), ActorID: actor, Action: "practice_registered", EntityType: "practice", EntityID: p.ID, Outcome: "success", RequestID: requestID, Details: p.CompanyName, CreatedAt: s.Clock.Now()})
}
func (s Practice) RecordHours(ctx context.Context, id string, h int, actor, requestID string) error {
	// Honour the request context up front: a cancelled or timed-out call must
	// not mutate practice data, so we never drop ctx in favour of
	// context.Background() as the previous detached write did.
	if e := ctx.Err(); e != nil {
		return e
	}
	p, e := s.Practices.Get(ctx, id)
	if e != nil {
		return e
	}
	if h < 0 || h > p.PlannedHours {
		return domain.ErrQuotaExceeded
	}
	now := s.Clock.Now()
	// Update the practice hours and the audit event inside a single
	// transaction bound to ctx. A late cancellation or timeout fails the
	// whole transaction (rollback) instead of leaving a written hours row
	// with a missing or independently-written audit record.
	tx, e := s.Practices.BeginTx(ctx)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	if e = s.Practices.UpdateHoursTx(ctx, tx, id, h, p.Version); e != nil {
		return e
	}
	ev := domain.AuditEvent{ID: token(), ActorID: actor, Action: "practice_hours_recorded", EntityType: "practice", EntityID: id, Outcome: "success", RequestID: requestID, Details: fmt.Sprintf("hours=%d", h), CreatedAt: now}
	if e = auditPracticeTx(ctx, tx, ev); e != nil {
		return e
	}
	return tx.Commit()
}
func auditPracticeTx(ctx context.Context, tx *sql.Tx, e domain.AuditEvent) error {
	_, err := tx.ExecContext(ctx, "INSERT INTO audit_events(id,actor_id,action,entity_type,entity_id,outcome,request_id,details,created_at) VALUES(?,?,?,?,?,?,?,?,?)", e.ID, e.ActorID, e.Action, e.EntityType, e.EntityID, e.Outcome, e.RequestID, e.Details, e.CreatedAt.Format(time.RFC3339Nano))
	return err
}
