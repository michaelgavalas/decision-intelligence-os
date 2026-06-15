package health

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// fakePinger returns the configured error from Ping.
type fakePinger struct {
	err error
}

func (f fakePinger) Ping(_ context.Context) error {
	return f.err
}

func decodeStatus(t *testing.T, body []byte) string {
	t.Helper()
	var resp statusResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("decode body %q: %v", body, err)
	}
	return resp.Status
}

func TestLiveness(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/livez", nil)

	Liveness().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := decodeStatus(t, rec.Body.Bytes()); got != "ok" {
		t.Errorf("body status = %q, want ok", got)
	}
}

func TestReadiness(t *testing.T) {
	tests := []struct {
		name       string
		pingErr    error
		wantCode   int
		wantStatus string
	}{
		{
			name:       "healthy",
			pingErr:    nil,
			wantCode:   http.StatusOK,
			wantStatus: "ok",
		},
		{
			name:       "unavailable",
			pingErr:    errors.New("connection refused"),
			wantCode:   http.StatusServiceUnavailable,
			wantStatus: "unavailable",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/readyz", nil)

			Readiness(fakePinger{err: tt.pingErr}).ServeHTTP(rec, req)

			if rec.Code != tt.wantCode {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantCode)
			}
			if got := decodeStatus(t, rec.Body.Bytes()); got != tt.wantStatus {
				t.Errorf("body status = %q, want %q", got, tt.wantStatus)
			}
		})
	}
}
