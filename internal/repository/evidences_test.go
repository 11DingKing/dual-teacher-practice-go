package repository

import (
	"context"
	"github.com/11DingKing/dual-teacher-practice-go/internal/domain"
	"github.com/11DingKing/dual-teacher-practice-go/internal/storage/sqlite"
	"testing"
	"time"
)

func evidenceDB(t *testing.T) *sqlite.DB { db := practiceDB(t); return db }
func TestEvidenceLifecycle(t *testing.T) {
	db := evidenceDB(t)
	r := Evidences{DB: db.SQL}
	e := domain.Evidence{ID: "e", ApplicationID: "a", Kind: "practice_log", Title: "日志", URI: "https://example/e", Status: domain.EvidencePending, SubmittedBy: "u", CreatedAt: time.Now().UTC()}
	if x := r.Add(context.Background(), e); x != nil {
		t.Fatal(x)
	}
	if x := r.SetStatus(context.Background(), "e", domain.EvidencePending, domain.EvidenceVerified, "u2", time.Now().UTC(), ""); x != nil {
		t.Fatal(x)
	}
	items, x := r.ForApplication(context.Background(), "a")
	if x != nil || len(items) != 1 || items[0].Status != domain.EvidenceVerified {
		t.Fatalf("%+v %v", items, x)
	}
}
func TestEvidenceWrongPreviousStatus(t *testing.T) {
	db := evidenceDB(t)
	r := Evidences{DB: db.SQL}
	e := domain.Evidence{ID: "e", ApplicationID: "a", Kind: "k", Title: "t", URI: "u", Status: domain.EvidencePending, SubmittedBy: "u", CreatedAt: time.Now().UTC()}
	if x := r.Add(context.Background(), e); x != nil {
		t.Fatal(x)
	}
	if x := r.SetStatus(context.Background(), "e", domain.EvidenceVerified, domain.EvidenceCorrection, "u2", time.Now().UTC(), "bad"); x != domain.ErrConflict {
		t.Fatalf("%v", x)
	}
}
