package service

import (
	"context"
	"testing"
	"time"

	"github.com/11DingKing/dual-teacher-practice-go/internal/clock"
	"github.com/11DingKing/dual-teacher-practice-go/internal/repository"
	"github.com/11DingKing/dual-teacher-practice-go/internal/storage/sqlite"
)

func TestRecordHoursHonorsCanceledContext(t *testing.T) {
	db, err := sqlite.Open(context.Background(), ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Date(2026, 8, 25, 0, 0, 0, 0, time.UTC)
	if _, err := db.SQL.Exec("INSERT INTO users VALUES('u','u','p','teacher','教师',NULL,'2026-01-01')"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.SQL.Exec("INSERT INTO quotas VALUES('q','Q',100,0,1000,0,1,1)"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.SQL.Exec("INSERT INTO applications VALUES('a','u','q',2026,'college_approved',10,10,'a',1,'2026-12-01','2026-01-01','2026-01-01')"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.SQL.Exec("INSERT INTO practices VALUES('p','a','企业','导师','mentor@example.com','planned',40,0,'2026-09-01','2026-10-01',1)"); err != nil {
		t.Fatal(err)
	}
	svc := Practice{Practices: repository.Practices{DB: db.SQL}, Apps: repository.Applications{DB: db.SQL}, Audit: repository.Audit{DB: db.SQL}, Clock: clock.Fixed{T: now}}
	if _, err := db.SQL.Exec("BEGIN IMMEDIATE"); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan error, 1)
	go func() { result <- svc.RecordHours(ctx, "p", 12, "u", "request") }()
	time.Sleep(100 * time.Millisecond)
	cancel()
	if _, err := db.SQL.Exec("COMMIT"); err != nil {
		t.Fatal(err)
	}
	if err := <-result; err == nil {
		t.Fatal("canceled request unexpectedly updated practice hours")
	}
	var actual int
	if err := db.SQL.QueryRow("SELECT actual_hours FROM practices WHERE id='p'").Scan(&actual); err != nil {
		t.Fatal(err)
	}
	if actual != 0 {
		t.Fatalf("canceled request changed actual hours to %d", actual)
	}
}
