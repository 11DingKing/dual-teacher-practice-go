package service

import (
	"github.com/11DingKing/dual-teacher-practice-go/internal/domain"
)

func Allowed(role domain.Role, action string) bool {
	switch role {
	case domain.RoleTeacher:
		return action == "create_application" || action == "submit_evidence" || action == "view_own"
	case domain.Role学院:
		return action == "approve_application" || action == "review_evidence" || action == "view_college"
	case domain.RoleEnterprise:
		return action == "review_practice" || action == "view_assigned"
	case domain.RoleHR:
		return action == "certify_application" || action == "renew_application" || action == "view_all"
	case domain.RoleAdmin:
		return true
	}
	return false
}
func Require(role domain.Role, action string) error {
	if !Allowed(role, action) {
		return domain.ErrForbidden
	}
	return nil
}
