package domain

import "errors"

var (
	ErrNotFound      = errors.New("not found")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrForbidden     = errors.New("forbidden")
	ErrConflict      = errors.New("conflict")
	ErrInvalidState  = errors.New("invalid state")
	ErrQuotaExceeded = errors.New("quota exceeded")
	ErrDeadline      = errors.New("deadline passed")
	ErrIdempotency   = errors.New("idempotency key conflict")
)
