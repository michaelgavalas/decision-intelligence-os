package httpx

import (
	"context"
	"net"
	"net/http"
	"strings"
)

type responseWriterKey struct{}

type requestKey struct{}

type clientIPKey struct{}

// HTTPContext stores the request's ResponseWriter and *http.Request in the
// request context so downstream code (notably GraphQL resolvers, which receive
// only a context.Context) can read and write transport-level details such as
// cookies. It must run before any handler that needs that access.
func HTTPContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), responseWriterKey{}, w)
		ctx = context.WithValue(ctx, requestKey{}, r)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ResponseWriter returns the http.ResponseWriter stored by HTTPContext, or
// ok=false when none is present.
func ResponseWriter(ctx context.Context) (http.ResponseWriter, bool) {
	w, ok := ctx.Value(responseWriterKey{}).(http.ResponseWriter)
	return w, ok
}

// Request returns the *http.Request stored by HTTPContext, or ok=false when
// none is present.
func Request(ctx context.Context) (*http.Request, bool) {
	r, ok := ctx.Value(requestKey{}).(*http.Request)
	return r, ok
}

// ClientIPMiddleware derives the client IP for the request and stores it in the
// context. It prefers the first hop of an X-Forwarded-For header when present
// and otherwise falls back to the host portion of RemoteAddr.
func ClientIPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := clientIPFrom(r)
		ctx := context.WithValue(r.Context(), clientIPKey{}, ip)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ClientIP returns the client IP stored by ClientIPMiddleware, or the empty
// string when none is present.
func ClientIP(ctx context.Context) string {
	if v, ok := ctx.Value(clientIPKey{}).(string); ok {
		return v
	}
	return ""
}

// clientIPFrom resolves the client IP, preferring the first X-Forwarded-For
// entry and falling back to the RemoteAddr host.
func clientIPFrom(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		first := strings.TrimSpace(strings.SplitN(xff, ",", 2)[0])
		if first != "" {
			return first
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return strings.TrimSpace(r.RemoteAddr)
	}
	return host
}
