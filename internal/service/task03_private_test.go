package service

import (
	"context"
	"github.com/11DingKing/dual-teacher-practice-go/internal/domain"
	"testing"
)

func TestEvidenceVerifyHonorsCanceledContext(t *testing.T) {
	e, _, db := evidenceFixture(t)
	if err := e.Add(context.Background(), domain.Evidence{ID: "e", ApplicationID: "a", Kind: "k", Title: "t", URI: "u"}, "u", "r"); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := e.Verify(ctx, "e", "r1", "req", true, "ok"); err == nil {
		t.Fatal("canceled verification accepted")
	}
	var status string
	_ = db.SQL.QueryRow("SELECT verification_status FROM evidences WHERE id='e'").Scan(&status)
	if status != "pending" {
		t.Fatalf("status changed: %s", status)
	}
}
