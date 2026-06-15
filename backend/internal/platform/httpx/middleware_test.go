package httpx

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/logging"
)

func TestRequestIDMiddlewareGeneratesID(t *testing.T) {
	var captured string
	handler := RequestIDMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		captured = RequestID(r.Context())
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rec, req)

	header := rec.Header().Get(RequestIDHeader)
	if header == "" {
		t.Fatal("expected generated X-Request-ID response header")
	}
	if captured == "" {
		t.Fatal("expected request id in context")
	}
	if captured != header {
		t.Errorf("context id %q != header id %q", captured, header)
	}
}

func TestRequestIDMiddlewarePreservesIncomingID(t *testing.T) {
	const incoming = "incoming-id-123"
	var captured string
	handler := RequestIDMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		captured = RequestID(r.Context())
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(RequestIDHeader, incoming)
	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get(RequestIDHeader); got != incoming {
		t.Errorf("response header = %q, want %q", got, incoming)
	}
	if captured != incoming {
		t.Errorf("context id = %q, want %q", captured, incoming)
	}
}

func TestRequestIDDefaultsToEmpty(t *testing.T) {
	if got := RequestID(context.Background()); got != "" {
		t.Errorf("RequestID = %q, want empty", got)
	}
}

func TestLoggerMiddlewareEmitsLine(t *testing.T) {
	var buf bytes.Buffer
	base := logging.New(&buf, "info")

	const incoming = "req-abc"
	// Chain request id then logger so the log line carries the request id.
	handler := RequestIDMiddleware(LoggerMiddleware(base)(http.HandlerFunc(
		func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusCreated)
		},
	)))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/widgets", nil)
	req.Header.Set(RequestIDHeader, incoming)
	handler.ServeHTTP(rec, req)

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("log line is not valid JSON: %v\n%s", err, buf.String())
	}

	if entry["request_id"] != incoming {
		t.Errorf("request_id = %v, want %q", entry["request_id"], incoming)
	}
	if status, ok := entry["status"].(float64); !ok || int(status) != http.StatusCreated {
		t.Errorf("status = %v, want %d", entry["status"], http.StatusCreated)
	}
	if entry["method"] != http.MethodPost {
		t.Errorf("method = %v, want %s", entry["method"], http.MethodPost)
	}
	if entry["path"] != "/widgets" {
		t.Errorf("path = %v, want /widgets", entry["path"])
	}
	if _, ok := entry["duration_ms"]; !ok {
		t.Error("expected duration_ms field")
	}
}

func TestLoggerDefaultsWhenAbsent(t *testing.T) {
	if Logger(context.Background()) == nil {
		t.Fatal("Logger returned nil, want default logger")
	}
}
