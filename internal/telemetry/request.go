package telemetry

import (
	"context"
	"sync/atomic"
	"time"
)

type RequestStats struct {
	started   atomic.Int64
	completed atomic.Int64
	failed    atomic.Int64
}

func (s *RequestStats) Begin()    { s.started.Add(1) }
func (s *RequestStats) Complete() { s.completed.Add(1) }
func (s *RequestStats) Fail()     { s.failed.Add(1) }
func (s *RequestStats) Snapshot() map[string]int64 {
	return map[string]int64{"started": s.started.Load(), "completed": s.completed.Load(), "failed": s.failed.Load()}
}
func WithDeadline(ctx context.Context, d time.Duration) (context.Context, context.CancelFunc) {
	if d <= 0 {
		d = time.Second
	}
	return context.WithTimeout(ctx, d)
}
