package service

import (
	"context"
	"fmt"
	"github.com/11DingKing/dual-teacher-practice-go/internal/clock"
	"github.com/11DingKing/dual-teacher-practice-go/internal/domain"
	"github.com/11DingKing/dual-teacher-practice-go/internal/repository"
)

type Certification struct {
	Apps     repository.Applications
	Reviews  repository.Reviews
	Evidence repository.Evidences
	Policy   domain.Policy
	Audit    repository.Audit
	Clock    clock.Clock
}

func (s Certification) Check(ctx context.Context, id string) (bool, string, error) {
	reviews, e := s.Reviews.ForApplication(ctx, id)
	if e != nil {
		return false, "", e
	}
	if len(reviews) < 2 {
		return false, "至少需要两名评价人", nil
	}
	for _, v := range reviews {
		if !s.Policy.ReviewPass(v.Score, v.Decision) {
			return false, "评价未达标", nil
		}
	}
	items, e := s.Evidence.ForApplication(ctx, id)
	if e != nil {
		return false, "", e
	}
	if len(items) == 0 {
		return false, "缺少证据", nil
	}
	for _, v := range items {
		if v.Status != domain.EvidenceVerified {
			return false, "存在未核验证据", nil
		}
	}
	return true, "", nil
}
func (s Certification) Approve(ctx context.Context, id, actor, request string) error {
	ok, reason, e := s.Check(ctx, id)
	if e != nil {
		return e
	}
	if !ok {
		return fmt.Errorf("%w: %s", domain.ErrConflict, reason)
	}
	a, e := s.Apps.Get(ctx, id)
	if e != nil {
		return e
	}
	if e = s.Apps.Transition(ctx, id, a.Status, domain.StatusCertified, s.Clock.Now()); e != nil {
		return e
	}
	return s.Audit.Append(ctx, domain.AuditEvent{ID: token(), ActorID: actor, Action: "certification_approved", EntityType: "application", EntityID: id, Outcome: "success", RequestID: request, Details: "policy passed", CreatedAt: s.Clock.Now()})
}
