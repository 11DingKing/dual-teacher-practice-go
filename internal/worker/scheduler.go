package worker

import (
	"context"
	"github.com/11DingKing/dual-teacher-practice-go/internal/domain"
	"github.com/11DingKing/dual-teacher-practice-go/internal/repository"
	"time"
)

type Scheduler struct {
	Queries  repository.Queries
	Jobs     repository.Jobs
	Interval time.Duration
}

func (s Scheduler) EnqueueDue(ctx context.Context, now time.Time) error {
	items, e := s.Queries.DueApplications(ctx, now, 100)
	if e != nil {
		return e
	}
	for _, a := range items {
		j := JobForApplication(a.ID, now)
		if e = s.Jobs.Enqueue(ctx, j); e != nil {
			return e
		}
	}
	return nil
}
func JobForApplication(id string, now time.Time) domain.WorkerJob {
	return domain.WorkerJob{ID: "due-" + id, JobType: "deadline_notice", Payload: id, Status: "pending", NextRunAt: now, CreatedAt: now, UpdatedAt: now}
}
