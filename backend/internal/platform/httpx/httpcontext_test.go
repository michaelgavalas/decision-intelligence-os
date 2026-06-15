package httpx

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPContextExposesWriterAndRequest(t *testing.T) {
	var (
		gotWriterOK  bool
		gotRequestOK bool
		storedHeader string
	)
	handler := HTTPContext(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		_, gotWriterOK = ResponseWriter(ctx)
		var stored *http.Request
		stored, gotRequestOK = Request(ctx)
		if stored != nil {
			storedHeader = stored.Header.Get("X-Probe")
		}
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Probe", "value-123")
	handler.ServeHTTP(rec, req)

	if !gotWriterOK {
		t.Error("ResponseWriter not available in context")
	}
	if !gotRequestOK {
		t.Error("Request not available in context")
	}
	if storedHeader != "value-123" {
		t.Errorf("stored request header = %q, want value-123", storedHeader)
	}

	// The stored ResponseWriter must be the one the middleware received so
	// resolvers can set headers and cookies through it.
	var wroteVia bool
	probe := HTTPContext(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		if w, ok := ResponseWriter(r.Context()); ok {
			w.Header().Set("X-From-Ctx", "yes")
			wroteVia = true
		}
	}))
	probeRec := httptest.NewRecorder()
	probe.ServeHTTP(probeRec, httptest.NewRequest(http.MethodGet, "/", nil))
	if !wroteVia {
		t.Fatal("expected to obtain ResponseWriter from context")
	}
	if probeRec.Header().Get("X-From-Ctx") != "yes" {
		t.Error("header set via context ResponseWriter did not reach the response")
	}
}

func TestHTTPContextAccessorsDefaultToFalse(t *testing.T) {
	if _, ok := ResponseWriter(context.Background()); ok {
		t.Error("ResponseWriter ok = true, want false on empty context")
	}
	if _, ok := Request(context.Background()); ok {
		t.Error("Request ok = true, want false on empty context")
	}
}

func TestClientIPPrefersForwardedForFirstHop(t *testing.T) {
	var captured string
	handler := ClientIPMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		captured = ClientIP(r.Context())
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:5555"
	req.Header.Set("X-Forwarded-For", "203.0.113.7, 70.41.3.18, 150.172.238.178")
	handler.ServeHTTP(rec, req)

	if captured != "203.0.113.7" {
		t.Errorf("ClientIP = %q, want 203.0.113.7", captured)
	}
}

func TestClientIPFallsBackToRemoteAddr(t *testing.T) {
	var captured string
	handler := ClientIPMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		captured = ClientIP(r.Context())
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "198.51.100.23:42000"
	handler.ServeHTTP(rec, req)

	if captured != "198.51.100.23" {
		t.Errorf("ClientIP = %q, want 198.51.100.23", captured)
	}
}

func TestClientIPDefaultsToEmpty(t *testing.T) {
	if got := ClientIP(context.Background()); got != "" {
		t.Errorf("ClientIP = %q, want empty", got)
	}
}
