package worker

import (
	"context"
	"encoding/json"
	"github.com/11DingKing/dual-teacher-practice-go/internal/domain"
	"github.com/11DingKing/dual-teacher-practice-go/internal/repository"
	"github.com/11DingKing/dual-teacher-practice-go/internal/telemetry"
	"sync"
	"time"
)

type Handler func(context.Context, domain.WorkerJob) error
type Worker struct {
	Jobs     repository.Jobs
	Interval time.Duration
	Logger   *telemetry.Logger
	Handler  Handler
	stop     chan struct{}
	once     sync.Once
}

func (w *Worker) Start(ctx context.Context) { w.stop = make(chan struct{}); go w.loop(ctx) }
func (w *Worker) Stop() {
	w.once.Do(func() {
		if w.stop != nil {
			close(w.stop)
		}
	})
}
func (w *Worker) loop(ctx context.Context) {
	ticker := time.NewTicker(w.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-w.stop:
			return
		case now := <-ticker.C:
			w.runOne(ctx, now)
		}
	}
}
func (w *Worker) runOne(ctx context.Context, now time.Time) {
	j, e := w.Jobs.Claim(ctx, now)
	if e != nil {
		return
	}
	if e = w.Handler(ctx, j); e != nil {
		next := now.Add(time.Duration(j.Attempts+1) * time.Second)
		_ = w.Jobs.Fail(ctx, j.ID, e.Error(), next, j.Attempts >= 5)
		w.Logger.Event("worker_failed", map[string]any{"job": j.ID, "attempt": j.Attempts})
	} else {
		_ = w.Jobs.Complete(ctx, j.ID)
		w.Logger.Event("worker_done", map[string]any{"job": j.ID})
	}
}
func MarshalPayload(v any) string { b, _ := json.Marshal(v); return string(b) }
