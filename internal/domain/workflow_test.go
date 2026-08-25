package domain

import "testing"

func TestWorkflowTransitions(t *testing.T) {
	cases := []struct {
		name     string
		from, to ApplicationStatus
		ok       bool
	}{
		{"draft submit", StatusDraft, StatusSubmitted, true},
		{"draft reject", StatusDraft, StatusRejected, true},
		{"submitted college", StatusSubmitted, StatusCollegeApproved, true},
		{"submitted reject", StatusSubmitted, StatusRejected, true},
		{"college evidence", StatusCollegeApproved, StatusEvidenceReview, true},
		{"evidence peer", StatusEvidenceReview, StatusPeerReview, true},
		{"peer certify", StatusPeerReview, StatusCertified, true},
		{"certified renewal", StatusCertified, StatusRenewal, true},
		{"certified expired", StatusCertified, StatusExpired, true},
		{"renewal certify", StatusRenewal, StatusCertified, true},
		{"renewal expired", StatusRenewal, StatusExpired, true},
		{"expired renewal", StatusExpired, StatusRenewal, true},
		{"rejected draft", StatusRejected, StatusDraft, true},
		{"draft peer invalid", StatusDraft, StatusPeerReview, false},
		{"submitted certify invalid", StatusSubmitted, StatusCertified, false},
		{"evidence draft invalid", StatusEvidenceReview, StatusDraft, false},
		{"peer draft invalid", StatusPeerReview, StatusDraft, false},
		{"expired submit invalid", StatusExpired, StatusSubmitted, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := CanTransition(tc.from, tc.to); got != tc.ok {
				t.Fatalf("got %v want %v", got, tc.ok)
			}
			err := Transition(tc.from, tc.to)
			if tc.ok && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !tc.ok && err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestWorkflowTerminalStates(t *testing.T) {
	for _, state := range []ApplicationStatus{StatusRejected, StatusExpired} {
		if !IsTerminal(state) {
			t.Fatalf("%s should be terminal", state)
		}
	}
	for _, state := range []ApplicationStatus{StatusDraft, StatusSubmitted, StatusCollegeApproved, StatusEvidenceReview, StatusPeerReview, StatusCertified, StatusRenewal} {
		if IsTerminal(state) {
			t.Fatalf("%s should not be terminal", state)
		}
	}
}

func TestTransitionSameStateIsIdempotent(t *testing.T) {
	for _, state := range []ApplicationStatus{StatusDraft, StatusSubmitted, StatusCertified, StatusExpired} {
		if err := Transition(state, state); err != nil {
			t.Fatalf("%s: %v", state, err)
		}
	}
}
