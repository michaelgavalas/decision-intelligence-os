package httpx

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

type loggerKey struct{}

// Logger returns the request-scoped logger stored in ctx, or slog.Default()
// when none is present.
func Logger(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(loggerKey{}).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}

// withLogger returns a copy of ctx carrying the given logger.
func withLogger(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, l)
}

// statusRecorder wraps http.ResponseWriter to capture the response status code.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

// WriteHeader records the status before delegating to the wrapped writer.
func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

// Write ensures a 200 status is recorded when the handler writes a body without
// an explicit WriteHeader call.
func (r *statusRecorder) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	return r.ResponseWriter.Write(b)
}

// LoggerMiddleware injects a request-scoped logger (the base logger annotated
// with the request id) into the context and emits one structured log line per
// request once the handler returns.
func LoggerMiddleware(base *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			requestID := RequestID(r.Context())
			logger := base.With(slog.String("request_id", requestID))

			ctx := withLogger(r.Context(), logger)
			rec := &statusRecorder{ResponseWriter: w}
			next.ServeHTTP(rec, r.WithContext(ctx))

			if rec.status == 0 {
				rec.status = http.StatusOK
			}
			logger.Info("request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", rec.status),
				slog.Int64("duration_ms", time.Since(start).Milliseconds()),
				slog.String("request_id", requestID),
			)
		})
	}
}
