package worker

import (
	"context"
	"errors"
	"github.com/11DingKing/dual-teacher-practice-go/internal/domain"
	"github.com/11DingKing/dual-teacher-practice-go/internal/repository"
	"github.com/11DingKing/dual-teacher-practice-go/internal/storage/sqlite"
	"github.com/11DingKing/dual-teacher-practice-go/internal/telemetry"
	"testing"
	"time"
)

func workerDB(t *testing.T) (repository.Jobs, *sqlite.DB) {
	db, e := sqlite.Open(context.Background(), ":memory:")
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { db.Close() })
	return repository.Jobs{DB: db.SQL}, db
}
func TestWorkerClaimComplete(t *testing.T) {
	jobs, db := workerDB(t)
	now := time.Now().UTC()
	j := domain.WorkerJob{ID: "j", JobType: "notice", Payload: "a", Status: "pending", NextRunAt: now.Add(-time.Second), CreatedAt: now, UpdatedAt: now}
	if e := jobs.Enqueue(context.Background(), j); e != nil {
		t.Fatal(e)
	}
	claimed, e := jobs.Claim(context.Background(), now)
	if e != nil || claimed.Status != "running" || claimed.Attempts != 1 {
		t.Fatalf("%+v %v", claimed, e)
	}
	if e = jobs.Complete(context.Background(), j.ID); e != nil {
		t.Fatal(e)
	}
	var status string
	if e = db.SQL.QueryRow("SELECT status FROM worker_jobs WHERE id='j'").Scan(&status); e != nil || status != "done" {
		t.Fatalf("%s %v", status, e)
	}
}
func TestWorkerClaimSkipsFuture(t *testing.T) {
	jobs, _ := workerDB(t)
	now := time.Now().UTC()
	if e := jobs.Enqueue(context.Background(), domain.WorkerJob{ID: "future", JobType: "n", Status: "pending", NextRunAt: now.Add(time.Hour), CreatedAt: now, UpdatedAt: now}); e != nil {
		t.Fatal(e)
	}
	if _, e := jobs.Claim(context.Background(), now); e == nil {
		t.Fatal("future job claimed")
	}
}
func TestWorkerFailureRetryAndPermanent(t *testing.T) {
	jobs, db := workerDB(t)
	now := time.Now().UTC()
	if e := jobs.Enqueue(context.Background(), domain.WorkerJob{ID: "j", JobType: "n", Status: "pending", NextRunAt: now, CreatedAt: now, UpdatedAt: now}); e != nil {
		t.Fatal(e)
	}
	j, e := jobs.Claim(context.Background(), now)
	if e != nil {
		t.Fatal(e)
	}
	if e = jobs.Fail(context.Background(), j.ID, "temporary", now.Add(time.Minute), false); e != nil {
		t.Fatal(e)
	}
	var status, next string
	if e = db.SQL.QueryRow("SELECT status,next_run_at FROM worker_jobs WHERE id='j'").Scan(&status, &next); e != nil || status != "pending" || next == "" {
		t.Fatalf("%s %s %v", status, next, e)
	}
	if e = jobs.Fail(context.Background(), j.ID, "fatal", now, true); e != nil {
		t.Fatal(e)
	}
	if e = db.SQL.QueryRow("SELECT status FROM worker_jobs WHERE id='j'").Scan(&status); e != nil || status != "failed" {
		t.Fatalf("%s %v", status, e)
	}
}
func TestWorkerRunOneCallsHandler(t *testing.T) {
	jobs, db := workerDB(t)
	now := time.Now().UTC()
	if e := jobs.Enqueue(context.Background(), domain.WorkerJob{ID: "j", JobType: "n", Status: "pending", NextRunAt: now, CreatedAt: now, UpdatedAt: now}); e != nil {
		t.Fatal(e)
	}
	called := false
	w := &Worker{Jobs: jobs, Logger: telemetry.New(), Handler: func(context.Context, domain.WorkerJob) error { called = true; return nil }}
	w.runOne(context.Background(), now)
	if !called {
		t.Fatal("handler not called")
	}
	var status string
	_ = db.SQL.QueryRow("SELECT status FROM worker_jobs WHERE id='j'").Scan(&status)
	if status != "done" {
		t.Fatal(status)
	}
}
func TestWorkerRunOneRecordsFailure(t *testing.T) {
	jobs, db := workerDB(t)
	now := time.Now().UTC()
	if e := jobs.Enqueue(context.Background(), domain.WorkerJob{ID: "j", JobType: "n", Status: "pending", NextRunAt: now, CreatedAt: now, UpdatedAt: now}); e != nil {
		t.Fatal(e)
	}
	w := &Worker{Jobs: jobs, Logger: telemetry.New(), Handler: func(context.Context, domain.WorkerJob) error { return errors.New("no") }}
	w.runOne(context.Background(), now)
	var status, last string
	if e := db.SQL.QueryRow("SELECT status,last_error FROM worker_jobs WHERE id='j'").Scan(&status, &last); e != nil || status != "pending" || last != "no" {
		t.Fatalf("%s %s %v", status, last, e)
	}
}
