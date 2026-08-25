package worker

import (
	"context"
	"errors"
	"github.com/11DingKing/dual-teacher-practice-go/internal/domain"
	"testing"
	"time"
)

func TestRunWithRetryEventuallySucceeds(t *testing.T) {
	attempts := 0
	fn := func(context.Context, domain.WorkerJob) error {
		attempts++
		if attempts < 3 {
			return errors.New("temporary")
		}
		return nil
	}
	n, e := RunWithRetry(context.Background(), domain.WorkerJob{}, fn)
	if e != nil || n != 3 || attempts != 3 {
		t.Fatalf("%d %d %v", n, attempts, e)
	}
}
func TestRunWithRetryPermanent(t *testing.T) {
	attempts := 0
	fn := func(context.Context, domain.WorkerJob) error { attempts++; return ErrPermanent }
	n, e := RunWithRetry(context.Background(), domain.WorkerJob{}, fn)
	if !errors.Is(e, ErrPermanent) || n != 1 || attempts != 1 {
		t.Fatalf("%d %d %v", n, attempts, e)
	}
}
func TestRunWithRetryCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	n, e := RunWithRetry(ctx, domain.WorkerJob{}, func(context.Context, domain.WorkerJob) error { return errors.New("temporary") })
	if !errors.Is(e, context.Canceled) || n != 0 {
		t.Fatalf("%d %v", n, e)
	}
}
func TestJobForApplication(t *testing.T) {
	now := time.Now().UTC()
	j := JobForApplication("a", now)
	if j.ID != "due-a" || j.JobType != "deadline_notice" || j.Payload != "a" || j.Status != "pending" {
		t.Fatalf("%+v", j)
	}
}
