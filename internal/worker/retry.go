package worker

import (
	"context"
	"errors"
	"github.com/11DingKing/dual-teacher-practice-go/internal/domain"
	"github.com/11DingKing/dual-teacher-practice-go/internal/repository"
	"time"
)

var ErrPermanent = errors.New("permanent worker failure")

func RunWithRetry(ctx context.Context, j domain.WorkerJob, fn Handler) (int, error) {
	attempt := j.Attempts
	for attempt < 6 {
		if e := ctx.Err(); e != nil {
			return attempt, e
		}
		attempt++
		if e := fn(ctx, j); e == nil {
			return attempt, nil
		} else if errors.Is(e, ErrPermanent) {
			return attempt, e
		} else {
			select {
			case <-ctx.Done():
				return attempt, ctx.Err()
			case <-time.After(repository.Backoff(attempt)):
			}
		}
	}
	return attempt, ErrPermanent
}
