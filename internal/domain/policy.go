package domain

import "time"

type Policy struct {
	MinPracticeHours int
	EvidenceKinds    []string
	ReviewPassScore  int
	RenewalWindow    time.Duration
}

func DefaultPolicy() Policy {
	return Policy{MinPracticeHours: 120, EvidenceKinds: []string{"practice_log", "service_contract", "course_plan", "evaluation"}, ReviewPassScore: 60, RenewalWindow: 90 * 24 * time.Hour}
}
func (p Policy) EvidenceKindAllowed(kind string) bool {
	for _, v := range p.EvidenceKinds {
		if v == kind {
			return true
		}
	}
	return false
}
func (p Policy) CanRenew(now, due time.Time, status ApplicationStatus) bool {
	return status == StatusCertified && !now.Before(due.Add(-p.RenewalWindow))
}
func (p Policy) ReviewPass(score int, decision ReviewDecision) bool {
	return decision == ReviewPass && score >= p.ReviewPassScore && score <= 100
}
