package domain

import "fmt"

func CanTransition(from, to ApplicationStatus) bool {
	allowed := map[ApplicationStatus]map[ApplicationStatus]bool{StatusDraft: {StatusSubmitted: true, StatusRejected: true}, StatusSubmitted: {StatusCollegeApproved: true, StatusRejected: true}, StatusCollegeApproved: {StatusEvidenceReview: true, StatusRejected: true}, StatusEvidenceReview: {StatusPeerReview: true, StatusRejected: true}, StatusPeerReview: {StatusCertified: true, StatusRejected: true}, StatusCertified: {StatusRenewal: true, StatusExpired: true}, StatusRenewal: {StatusCertified: true, StatusExpired: true}, StatusRejected: {StatusDraft: true}, StatusExpired: {StatusRenewal: true}}
	return allowed[from][to]
}
func Transition(from, to ApplicationStatus) error {
	if from == to {
		return nil
	}
	if !CanTransition(from, to) {
		return fmt.Errorf("%w: %s -> %s", ErrInvalidState, from, to)
	}
	return nil
}
func IsTerminal(s ApplicationStatus) bool { return s == StatusRejected || s == StatusExpired }
