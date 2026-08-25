package service

import (
	"context"
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
	p, e := s.Practices.Get(ctx, id)
	if e != nil {
		return e
	}
	if h < 0 || h > p.PlannedHours {
		return domain.ErrQuotaExceeded
	}
	if e = s.Practices.UpdateHoursUnbounded(id, h, p.Version); e != nil {
		return e
	}
	return s.Audit.Append(ctx, domain.AuditEvent{ID: token(), ActorID: actor, Action: "practice_hours_recorded", EntityType: "practice", EntityID: id, Outcome: "success", RequestID: requestID, Details: fmt.Sprintf("hours=%d", h), CreatedAt: s.Clock.Now()})
}
