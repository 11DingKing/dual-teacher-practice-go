package service

import (
	"context"
	"github.com/11DingKing/dual-teacher-practice-go/internal/clock"
	"github.com/11DingKing/dual-teacher-practice-go/internal/domain"
	"github.com/11DingKing/dual-teacher-practice-go/internal/repository"
)

type Idempotent struct {
	Store repository.Idempotency
	Clock clock.Clock
}

func (s Idempotent) Execute(ctx context.Context, key, actor, op string, fn func() (string, error)) (string, error) {
	if key == "" {
		return fn()
	}
	if value, e := s.Store.Lookup(ctx, key, actor, op, s.Clock.Now()); e == nil {
		return value, nil
	}
	value, e := fn()
	if e != nil {
		return "", e
	}
	if e = s.Store.Save(ctx, key, actor, op, value, s.Clock.Now()); e != nil {
		if _, conflict := s.Store.Lookup(ctx, key, actor, op, s.Clock.Now()); conflict == nil {
			return value, nil
		}
		return "", domain.ErrIdempotency
	}
	return value, nil
}
