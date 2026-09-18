package httpx

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/As71er/once/internal/apperrors"
)

type errorResponse struct {
	Error string `json:"error"`
}

func describe(err error) (status int, clientMsg string) {
	switch {
	case errors.Is(err, apperrors.ErrNotFound):
		return http.StatusNotFound, "Resource not found"
	case errors.Is(err, apperrors.ErrInvalidInput):
		return http.StatusBadRequest, err.Error()
	case errors.Is(err, context.DeadlineExceeded):
		return http.StatusGatewayTimeout, "Request timed out"
	case errors.Is(err, context.Canceled):
		return 499, "Request canceled"
	default:
		return http.StatusInternalServerError, "Something went wrong"
	}
}

func Error(w http.ResponseWriter, r *http.Request, err error) {
	status, msg := describe(err)
	LoggerFrom(r.Context()).Warn("bad request", "status", status, "err", err)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(errorResponse{Error: msg})
}
