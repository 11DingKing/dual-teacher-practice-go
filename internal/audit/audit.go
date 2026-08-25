package audit

import (
	"context"
	"fmt"
	"github.com/11DingKing/dual-teacher-practice-go/internal/clock"
	"github.com/11DingKing/dual-teacher-practice-go/internal/domain"
	"github.com/11DingKing/dual-teacher-practice-go/internal/repository"
	"time"
)

type Service struct {
	Repo  repository.Audit
	Clock clock.Clock
}

func (s Service) Record(ctx context.Context, actor, action, typ, id, outcome, request, details string) error {
	return s.Repo.Append(ctx, domain.AuditEvent{ID: randomID(), ActorID: actor, Action: action, EntityType: typ, EntityID: id, Outcome: outcome, RequestID: request, Details: details, CreatedAt: s.Clock.Now()})
}
func randomID() string { return fmt.Sprintf("audit-%d", time.Now().UnixNano()) }
