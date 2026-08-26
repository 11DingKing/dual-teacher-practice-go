package service

import (
	"github.com/11DingKing/dual-teacher-practice-go/internal/domain"
	"testing"
)

func TestAllowedRoleMatrix(t *testing.T) {
	cases := []struct {
		role   domain.Role
		action string
		ok     bool
	}{
		{domain.RoleTeacher, "create_application", true}, {domain.RoleTeacher, "submit_evidence", true}, {domain.RoleTeacher, "view_own", true}, {domain.RoleTeacher, "approve_application", false}, {domain.RoleTeacher, "certify_application", false}, {domain.RoleTeacher, "view_all", false},
		{domain.Role学院, "approve_application", true}, {domain.Role学院, "review_evidence", true}, {domain.Role学院, "view_college", true}, {domain.Role学院, "create_application", false}, {domain.Role学院, "certify_application", false}, {domain.Role学院, "renew_application", false},
		{domain.RoleEnterprise, "review_practice", true}, {domain.RoleEnterprise, "view_assigned", true}, {domain.RoleEnterprise, "approve_application", false}, {domain.RoleEnterprise, "view_all", false}, {domain.RoleEnterprise, "renew_application", false},
		{domain.RoleHR, "certify_application", true}, {domain.RoleHR, "renew_application", true}, {domain.RoleHR, "view_all", true}, {domain.RoleHR, "create_application", false}, {domain.RoleHR, "review_practice", false},
		{domain.RoleAdmin, "anything", true}, {domain.RoleAdmin, "delete_everything", true}, {domain.RoleAdmin, "view_all", true}, {domain.Role("unknown"), "anything", false},
	}
	for i, tc := range cases {
		t.Run(string(tc.role)+"-"+tc.action, func(t *testing.T) {
			if got := Allowed(tc.role, tc.action); got != tc.ok {
				t.Fatalf("case %d got %v", i, got)
			}
			err := Require(tc.role, tc.action)
			if tc.ok && err != nil {
				t.Fatal(err)
			}
			if !tc.ok && err == nil {
				t.Fatal("expected forbidden")
			}
		})
	}
}
func TestAllowedDoesNotMatchPrefixes(t *testing.T) {
	for _, action := range []string{"create_application_extra", "view_own_all", "approve_application/1", ""} {
		if Allowed(domain.RoleTeacher, action) {
			t.Fatal(action)
		}
	}
}
func TestAdminAlwaysAllowed(t *testing.T) {
	for _, action := range []string{"a", "b", "certify_application", "", "unusual"} {
		if !Allowed(domain.RoleAdmin, action) {
			t.Fatal(action)
		}
	}
}
