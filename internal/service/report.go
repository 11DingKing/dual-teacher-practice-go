package service

import (
	"context"
	"github.com/11DingKing/dual-teacher-practice-go/internal/domain"
	"github.com/11DingKing/dual-teacher-practice-go/internal/pagination"
	"github.com/11DingKing/dual-teacher-practice-go/internal/repository"
)

type Report struct {
	Apps  repository.Applications
	Audit repository.Audit
}

func (s Report) List(ctx context.Context, status string, p pagination.Page) ([]domain.Application, error) {
	return s.Apps.List(ctx, status, p.Limit, p.Offset)
}
func (s Report) Timeline(ctx context.Context, id string) ([]domain.AuditEvent, error) {
	return s.Audit.ForEntity(ctx, "application", id)
}
