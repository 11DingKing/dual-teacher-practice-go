package service

import (
	"context"
	"fmt"
	"github.com/11DingKing/dual-teacher-practice-go/internal/clock"
	"github.com/11DingKing/dual-teacher-practice-go/internal/domain"
	"github.com/11DingKing/dual-teacher-practice-go/internal/repository"
)

type Review struct {
	Reviews   repository.Reviews
	Evidences repository.Evidences
	Apps      repository.Applications
	Audit     repository.Audit
	Clock     clock.Clock
}

func (s Review) Add(ctx context.Context, v domain.Review, actor, requestID string) error {
	a, e := s.Apps.Get(ctx, v.ApplicationID)
	if e != nil {
		return e
	}
	if a.Status != domain.StatusPeerReview {
		return domain.ErrInvalidState
	}
	if v.Score < 0 || v.Score > 100 {
		return fmt.Errorf("%w: score", domain.ErrConflict)
	}
	v.ReviewerID = actor
	v.CreatedAt = s.Clock.Now()
	if e = s.Reviews.Add(ctx, v); e != nil {
		return e
	}
	return s.Audit.Append(ctx, domain.AuditEvent{ID: token(), ActorID: actor, Action: "review_added", EntityType: "application", EntityID: v.ApplicationID, Outcome: "success", RequestID: requestID, Details: string(v.Decision), CreatedAt: s.Clock.Now()})
}
func (s Review) Certify(ctx context.Context, id, actor, requestID string) error {
	reviews, e := s.Reviews.ForApplication(ctx, id)
	if e != nil {
		return e
	}
	if len(reviews) < 2 {
		return fmt.Errorf("%w: two reviews required", domain.ErrConflict)
	}
	for _, r := range reviews {
		if r.Decision != domain.ReviewPass {
			return fmt.Errorf("%w: review failed", domain.ErrConflict)
		}
	}
	ev, e := s.Evidences.ForApplication(ctx, id)
	if e != nil {
		return e
	}
	if len(ev) == 0 {
		return fmt.Errorf("%w: evidence required", domain.ErrConflict)
	}
	for _, v := range ev {
		if v.Status != domain.EvidenceVerified {
			return fmt.Errorf("%w: evidence pending", domain.ErrConflict)
		}
	}
	a, e := s.Apps.Get(ctx, id)
	if e != nil {
		return e
	}
	if e = s.Apps.Transition(ctx, id, a.Status, domain.StatusCertified, s.Clock.Now()); e != nil {
		return e
	}
	return s.Audit.Append(ctx, domain.AuditEvent{ID: token(), ActorID: actor, Action: "application_certified", EntityType: "application", EntityID: id, Outcome: "success", RequestID: requestID, Details: "two reviews and verified evidence", CreatedAt: s.Clock.Now()})
}
