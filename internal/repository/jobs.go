package repository

import (
	"context"
	"database/sql"
	"github.com/11DingKing/dual-teacher-practice-go/internal/domain"
	"time"
)

type Jobs struct{ DB *sql.DB }

func (r Jobs) Enqueue(ctx context.Context, j domain.WorkerJob) error {
	_, e := r.DB.ExecContext(ctx, "INSERT INTO worker_jobs(id,job_type,payload,status,attempts,next_run_at,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?)", j.ID, j.JobType, j.Payload, j.Status, j.Attempts, j.NextRunAt.Format(time.RFC3339Nano), j.CreatedAt.Format(time.RFC3339Nano), j.UpdatedAt.Format(time.RFC3339Nano))
	return e
}
func (r Jobs) Claim(ctx context.Context, now time.Time) (domain.WorkerJob, error) {
	tx, e := r.DB.BeginTx(ctx, nil)
	if e != nil {
		return domain.WorkerJob{}, e
	}
	defer tx.Rollback()
	var j domain.WorkerJob
	var next, created, updated string
	e = tx.QueryRowContext(ctx, "SELECT id,job_type,payload,status,attempts,next_run_at,created_at,updated_at FROM worker_jobs WHERE status='pending' AND next_run_at<=? ORDER BY next_run_at LIMIT 1", now.Format(time.RFC3339Nano)).Scan(&j.ID, &j.JobType, &j.Payload, &j.Status, &j.Attempts, &next, &created, &updated)
	if e != nil {
		return j, e
	}
	res, e := tx.ExecContext(ctx, "UPDATE worker_jobs SET status='running',attempts=attempts+1,updated_at=? WHERE id=? AND status='pending'", now.Format(time.RFC3339Nano), j.ID)
	if e != nil {
		return j, e
	}
	n, _ := res.RowsAffected()
	if n != 1 {
		return j, domain.ErrConflict
	}
	if e = tx.Commit(); e != nil {
		return j, e
	}
	j.Status = "running"
	j.Attempts++
	j.NextRunAt, _ = time.Parse(time.RFC3339Nano, next)
	j.CreatedAt, _ = time.Parse(time.RFC3339Nano, created)
	j.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updated)
	return j, nil
}
func (r Jobs) Complete(ctx context.Context, id string) error {
	_, e := r.DB.ExecContext(ctx, "UPDATE worker_jobs SET status='done',updated_at=? WHERE id=?", time.Now().UTC().Format(time.RFC3339Nano), id)
	return e
}
func (r Jobs) Fail(ctx context.Context, id, msg string, next time.Time, permanent bool) error {
	st := "pending"
	if permanent {
		st = "failed"
	}
	_, e := r.DB.ExecContext(ctx, "UPDATE worker_jobs SET status=?,last_error=?,next_run_at=?,updated_at=? WHERE id=?", st, msg, next.Format(time.RFC3339Nano), time.Now().UTC().Format(time.RFC3339Nano), id)
	return e
}
