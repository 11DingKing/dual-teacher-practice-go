package httpapi

import (
	"encoding/json"
	"errors"
	"github.com/11DingKing/dual-teacher-practice-go/internal/domain"
	"net/http"
)

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	code := "internal_error"
	if errors.Is(err, domain.ErrUnauthorized) {
		status = 401
		code = "unauthorized"
	}
	if errors.Is(err, domain.ErrForbidden) {
		status = 403
		code = "forbidden"
	}
	if errors.Is(err, domain.ErrNotFound) {
		status = 404
		code = "not_found"
	}
	if errors.Is(err, domain.ErrConflict) || errors.Is(err, domain.ErrQuotaExceeded) {
		status = 409
		code = "conflict"
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(apiError{Code: code, Message: err.Error()})
}
