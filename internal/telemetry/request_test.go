package telemetry

import (
	"context"
	"testing"
	"time"
)

func TestRequestStats(t *testing.T) {
	s := &RequestStats{}
	s.Begin()
	s.Begin()
	s.Complete()
	s.Fail()
	v := s.Snapshot()
	if v["started"] != 2 || v["completed"] != 1 || v["failed"] != 1 {
		t.Fatalf("%v", v)
	}
}
func TestRequestStatsConcurrent(t *testing.T) {
	s := &RequestStats{}
	done := make(chan struct{})
	for i := 0; i < 20; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				s.Begin()
				s.Complete()
			}
			done <- struct{}{}
		}()
	}
	for i := 0; i < 20; i++ {
		<-done
	}
	v := s.Snapshot()
	if v["started"] != 2000 || v["completed"] != 2000 {
		t.Fatalf("%v", v)
	}
}
func TestWithDeadline(t *testing.T) {
	ctx, cancel := WithDeadline(context.Background(), time.Millisecond)
	defer cancel()
	select {
	case <-ctx.Done():
	case <-time.After(time.Second):
		t.Fatal("deadline did not fire")
	}
}
func TestWithDeadlineFallback(t *testing.T) {
	ctx, cancel := WithDeadline(context.Background(), 0)
	defer cancel()
	if _, ok := ctx.Deadline(); !ok {
		t.Fatal("deadline missing")
	}
}
