package domain

import (
	"net/mail"
	"strings"
	"time"
)

func ValidateApplication(a Application) error {
	if strings.TrimSpace(a.ID) == "" || strings.TrimSpace(a.TeacherID) == "" || strings.TrimSpace(a.QuotaID) == "" {
		return ErrConflict
	}
	if a.Year < 2020 || a.Year > time.Now().UTC().Year()+1 {
		return ErrConflict
	}
	if a.RequestedHours <= 0 || a.RequestedHours > 2000 {
		return ErrConflict
	}
	if a.RequestedBudgetCents < 0 {
		return ErrConflict
	}
	return nil
}
func ValidatePractice(p Practice) error {
	if strings.TrimSpace(p.CompanyName) == "" || strings.TrimSpace(p.MentorName) == "" {
		return ErrConflict
	}
	if _, e := mail.ParseAddress(p.ContactEmail); e != nil {
		return ErrConflict
	}
	if p.PlannedHours <= 0 || !p.EndsAt.After(p.StartsAt) {
		return ErrConflict
	}
	return nil
}
func ValidateEvidence(e Evidence) error {
	if strings.TrimSpace(e.Title) == "" || strings.TrimSpace(e.URI) == "" || strings.TrimSpace(e.Kind) == "" {
		return ErrConflict
	}
	return nil
}
func ValidateReview(r Review) error {
	if r.Score < 0 || r.Score > 100 || strings.TrimSpace(r.Comment) == "" {
		return ErrConflict
	}
	return nil
}
