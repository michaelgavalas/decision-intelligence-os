// Package httpx provides HTTP middleware shared across the backend: request id
// propagation and request-scoped structured logging.
package httpx

import (
	"context"
	"net/http"

	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/id"
)

// RequestIDHeader is the canonical header used to carry the request id on both
// inbound requests and outbound responses.
const RequestIDHeader = "X-Request-ID"

type requestIDKey struct{}

// RequestID returns the request id stored in ctx, or the empty string when
// none is present.
func RequestID(ctx context.Context) string {
	if v, ok := ctx.Value(requestIDKey{}).(string); ok {
		return v
	}
	return ""
}

// withRequestID returns a copy of ctx carrying the given request id.
func withRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, requestID)
}

// RequestIDMiddleware reads the inbound X-Request-ID header, generating a fresh
// UUIDv7 when it is absent, stores the value in the request context, and echoes
// it on the response.
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get(RequestIDHeader)
		if requestID == "" {
			requestID = id.New().String()
		}
		w.Header().Set(RequestIDHeader, requestID)
		ctx := withRequestID(r.Context(), requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
