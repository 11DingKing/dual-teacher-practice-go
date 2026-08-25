package domain

import (
	"testing"
	"time"
)

func TestDefaultPolicyEvidenceKinds(t *testing.T) {
	p := DefaultPolicy()
	for _, kind := range []string{"practice_log", "service_contract", "course_plan", "evaluation"} {
		if !p.EvidenceKindAllowed(kind) {
			t.Fatalf("kind %s rejected", kind)
		}
	}
	for _, kind := range []string{"paper_count", "unknown", ""} {
		if p.EvidenceKindAllowed(kind) {
			t.Fatalf("kind %s accepted", kind)
		}
	}
}
func TestPolicyRenewalWindow(t *testing.T) {
	p := DefaultPolicy()
	due := time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)
	before := due.Add(-p.RenewalWindow).Add(-time.Second)
	at := due.Add(-p.RenewalWindow)
	after := due.Add(-p.RenewalWindow).Add(time.Second)
	if p.CanRenew(before, due, StatusCertified) {
		t.Fatal("before window")
	}
	if !p.CanRenew(at, due, StatusCertified) {
		t.Fatal("at window")
	}
	if !p.CanRenew(after, due, StatusCertified) {
		t.Fatal("after window")
	}
	for _, st := range []ApplicationStatus{StatusDraft, StatusSubmitted, StatusRenewal, StatusExpired} {
		if p.CanRenew(after, due, st) {
			t.Fatalf("state %s accepted", st)
		}
	}
}
func TestReviewPolicy(t *testing.T) {
	p := DefaultPolicy()
	for _, score := range []int{60, 70, 100} {
		if !p.ReviewPass(score, ReviewPass) {
			t.Fatal(score)
		}
	}
	for _, score := range []int{-1, 0, 59, 101} {
		if p.ReviewPass(score, ReviewPass) {
			t.Fatal(score)
		}
	}
	if p.ReviewPass(100, ReviewFail) {
		t.Fatal("failed decision")
	}
}
func TestValidation(t *testing.T) {
	now := time.Now().UTC()
	valid := Application{ID: "a", TeacherID: "t", QuotaID: "q", Year: now.Year(), RequestedHours: 1}
	if e := ValidateApplication(valid); e != nil {
		t.Fatal(e)
	}
	invalid := []Application{{}, {ID: "a", TeacherID: "t", QuotaID: "q", Year: 2010, RequestedHours: 1}, {ID: "a", TeacherID: "t", QuotaID: "q", Year: now.Year(), RequestedHours: 0}, {ID: "a", TeacherID: "t", QuotaID: "q", Year: now.Year(), RequestedHours: 3000}}
	for i, v := range invalid {
		if e := ValidateApplication(v); e == nil {
			t.Fatalf("invalid %d accepted", i)
		}
	}
}
func TestPracticeValidation(t *testing.T) {
	start := time.Now().UTC()
	base := Practice{CompanyName: "企业", MentorName: "导师", ContactEmail: "mentor@example.com", PlannedHours: 10, StartsAt: start, EndsAt: start.Add(time.Hour)}
	if e := ValidatePractice(base); e != nil {
		t.Fatal(e)
	}
	cases := []Practice{{}, base, {CompanyName: "企业", MentorName: "导师", ContactEmail: "bad", PlannedHours: 10, StartsAt: start, EndsAt: start.Add(time.Hour)}, {CompanyName: "企业", MentorName: "导师", ContactEmail: "a@b.com", PlannedHours: 0, StartsAt: start, EndsAt: start.Add(time.Hour)}, {CompanyName: "企业", MentorName: "导师", ContactEmail: "a@b.com", PlannedHours: 10, StartsAt: start, EndsAt: start}}
	for i, v := range cases {
		if i == 1 {
			continue
		}
		if e := ValidatePractice(v); e == nil {
			t.Fatalf("case %d accepted", i)
		}
	}
}
