package service

import (
	"context"
	"testing"
	"time"

	"github.com/11DingKing/dual-teacher-practice-go/internal/clock"
	"github.com/11DingKing/dual-teacher-practice-go/internal/domain"
	"github.com/11DingKing/dual-teacher-practice-go/internal/repository"
	"github.com/11DingKing/dual-teacher-practice-go/internal/storage/sqlite"
)

func TestApplicationSubmitRollsBackWhenAuditWriteFails(t *testing.T) {
	db, err := sqlite.Open(context.Background(), ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Date(2026, 8, 25, 0, 0, 0, 0, time.UTC)
	users := repository.Users{DB: db.SQL}
	if err := users.Create(context.Background(), domain.User{ID: "teacher", Username: "teacher", DisplayName: "教师", Role: domain.RoleTeacher, CreatedAt: now}, "pw"); err != nil {
		t.Fatal(err)
	}
	if err := (repository.Quotas{DB: db.SQL}).Create(context.Background(), domain.Quota{ID: "quota", Name: "年度实践", MaxHours: 100, MaxBudgetCents: 10000, Version: 1, Active: true}); err != nil {
		t.Fatal(err)
	}
	svc := Application{DB: db.SQL, Apps: repository.Applications{DB: db.SQL}, Quotas: repository.Quotas{DB: db.SQL}, Audit: repository.Audit{DB: db.SQL}, Clock: clock.Fixed{T: now}}
	_, err = svc.Submit(context.Background(), domain.Application{ID: "application", TeacherID: "teacher", QuotaID: "quota", Year: 2026, RequestedHours: 20, RequestedBudgetCents: 500, IdempotencyKey: "idem"}, "missing-actor", "request-1")
	if err == nil {
		t.Fatal("audit failure should be returned")
	}
	var applications, usedHours int
	if err := db.SQL.QueryRow("SELECT COUNT(*) FROM applications").Scan(&applications); err != nil {
		t.Fatal(err)
	}
	if err := db.SQL.QueryRow("SELECT used_hours FROM quotas WHERE id='quota'").Scan(&usedHours); err != nil {
		t.Fatal(err)
	}
	if applications != 0 || usedHours != 0 {
		t.Fatalf("transaction leaked application=%d used_hours=%d", applications, usedHours)
	}
}
