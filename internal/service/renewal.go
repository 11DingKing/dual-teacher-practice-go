package service

import (
	"context"
	"fmt"
	"github.com/11DingKing/dual-teacher-practice-go/internal/clock"
	"github.com/11DingKing/dual-teacher-practice-go/internal/domain"
	"github.com/11DingKing/dual-teacher-practice-go/internal/repository"
)

type Renewal struct {
	Apps      repository.Applications
	Practices repository.Practices
	Policy    domain.Policy
	Clock     clock.Clock
	Audit     repository.Audit
}

func (s Renewal) Start(ctx context.Context, id, actor, request string) error {
	a, e := s.Apps.Get(ctx, id)
	if e != nil {
		return e
	}
	if !s.Policy.CanRenew(s.Clock.Now(), a.DueAt, a.Status) {
		return fmt.Errorf("%w: renewal window", domain.ErrInvalidState)
	}
	if e = s.Apps.Transition(ctx, id, a.Status, domain.StatusRenewal, s.Clock.Now()); e != nil {
		return e
	}
	return s.Audit.Append(ctx, domain.AuditEvent{ID: token(), ActorID: actor, Action: "renewal_started", EntityType: "application", EntityID: id, Outcome: "success", RequestID: request, Details: "renewal window", CreatedAt: s.Clock.Now()})
}
