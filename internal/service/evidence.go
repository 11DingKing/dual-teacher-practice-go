package service

import (
	"context"
	"fmt"
	"github.com/11DingKing/dual-teacher-practice-go/internal/clock"
	"github.com/11DingKing/dual-teacher-practice-go/internal/domain"
	"github.com/11DingKing/dual-teacher-practice-go/internal/repository"
)

type Evidence struct {
	Evidence repository.Evidences
	Apps     repository.Applications
	Audit    repository.Audit
	Clock    clock.Clock
}

func (s Evidence) Add(ctx context.Context, e domain.Evidence, actor, requestID string) error {
	a, err := s.Apps.Get(ctx, e.ApplicationID)
	if err != nil {
		return err
	}
	if a.Status != domain.StatusEvidenceReview && a.Status != domain.StatusCollegeApproved {
		return fmt.Errorf("%w: evidence window", domain.ErrInvalidState)
	}
	e.Status = domain.EvidencePending
	e.SubmittedBy = actor
	e.CreatedAt = s.Clock.Now()
	if err = s.Evidence.Add(ctx, e); err != nil {
		return err
	}
	return s.Audit.Append(ctx, domain.AuditEvent{ID: token(), ActorID: actor, Action: "evidence_submitted", EntityType: "application", EntityID: e.ApplicationID, Outcome: "success", RequestID: requestID, Details: e.Kind, CreatedAt: s.Clock.Now()})
}
func (s Evidence) Verify(ctx context.Context, id, actor, requestID string, approved bool, note string) error {
	to := domain.EvidenceCorrection
	if approved {
		to = domain.EvidenceVerified
	}
	if err := s.Evidence.SetStatus(ctx, id, domain.EvidencePending, to, actor, s.Clock.Now(), note); err != nil {
		return err
	}
	return s.Audit.Append(ctx, domain.AuditEvent{ID: token(), ActorID: actor, Action: "evidence_reviewed", EntityType: "evidence", EntityID: id, Outcome: "success", RequestID: requestID, Details: fmt.Sprintf("approved=%t", approved), CreatedAt: s.Clock.Now()})
}
