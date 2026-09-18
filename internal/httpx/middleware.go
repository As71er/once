package httpx

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/As71er/once/internal/apperrors"
)

type ctxKey int

const reqLoggerKey ctxKey = iota

func LoggerFrom(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(reqLoggerKey).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := fmt.Sprintf("%d", time.Now().UnixNano()) // TODO: temp, change to UUID

		logger := slog.With(
			"request_id", reqID,
			"method", r.Method,
			"path", r.URL.Path,
		)

		w.Header().Set("X-Request-ID", reqID)

		ctx := context.WithValue(r.Context(), reqLoggerKey, logger)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				LoggerFrom(r.Context()).Error("panic recovered",
					"panic", rec,
					"stack", debug.Stack(),
				)

				Error(w, r, fmt.Errorf("internal: %w", apperrors.ErrInternal))
			}
		}()
		next.ServeHTTP(w, r.WithContext(r.Context()))
	})
}
