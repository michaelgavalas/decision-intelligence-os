package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/config"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/dbtest"
)

// newTestApp builds an App against a freshly migrated container database with
// development defaults.
func newTestApp(t *testing.T) *App {
	t.Helper()

	_, dsn := dbtest.NewPoolWithURL(t)

	// Load development defaults, supplying only the database URL.
	cfg, err := config.Load(func(k string) string {
		if k == "DATABASE_URL" {
			return dsn
		}
		return ""
	})
	if err != nil {
		t.Fatalf("config.Load: %v", err)
	}

	app, err := NewApp(context.Background(), cfg, testLogger())
	if err != nil {
		t.Fatalf("NewApp: %v", err)
	}
	t.Cleanup(app.Close)
	return app
}

// postGraphQL sends a GraphQL query and returns the decoded response.
func postGraphQL(t *testing.T, h http.Handler, query string) map[string]any {
	t.Helper()

	body, err := json.Marshal(map[string]string{"query": query})
	if err != nil {
		t.Fatalf("marshal query: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/graphql", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("graphql status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}

	var out map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode response: %v; body=%s", err, rec.Body.String())
	}
	return out
}

func TestApp_Healthz(t *testing.T) {
	app := newTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	app.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("healthz status = %d, want 200", rec.Code)
	}
}

func TestApp_SecurityHeaders(t *testing.T) {
	app := newTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	app.Handler.ServeHTTP(rec, req)

	want := map[string]string{
		"X-Content-Type-Options":  "nosniff",
		"X-Frame-Options":         "DENY",
		"Referrer-Policy":         "strict-origin-when-cross-origin",
		"Permissions-Policy":      "geolocation=(), microphone=(), camera=()",
		"Content-Security-Policy": "default-src 'none'; frame-ancestors 'none'",
	}
	for header, value := range want {
		if got := rec.Header().Get(header); got != value {
			t.Errorf("%s = %q, want %q", header, got, value)
		}
	}

	// HSTS must be gated on secure cookies; development defaults leave it off.
	if hsts := rec.Header().Get("Strict-Transport-Security"); hsts != "" {
		t.Errorf("Strict-Transport-Security = %q, want empty without CookieSecure", hsts)
	}
}

func TestSecurityHeadersHSTSGatedOnSecure(t *testing.T) {
	t.Run("secure on", func(t *testing.T) {
		rec := serveSecurityHeaders(t, config.Config{CookieSecure: true})
		if got := rec.Header().Get("Strict-Transport-Security"); got != "max-age=31536000; includeSubDomains" {
			t.Errorf("HSTS = %q, want max-age=31536000; includeSubDomains", got)
		}
	})
	t.Run("secure off", func(t *testing.T) {
		rec := serveSecurityHeaders(t, config.Config{CookieSecure: false})
		if got := rec.Header().Get("Strict-Transport-Security"); got != "" {
			t.Errorf("HSTS = %q, want empty", got)
		}
	})
}

// serveSecurityHeaders runs a no-op handler through the securityHeaders
// middleware built from cfg and returns the recorder.
func serveSecurityHeaders(t *testing.T, cfg config.Config) *httptest.ResponseRecorder {
	t.Helper()
	h := securityHeaders(cfg)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	return rec
}

func TestApp_RegisterMutation(t *testing.T) {
	app := newTestApp(t)

	const mutation = `mutation { register(input:{email:"a@b.com", name:"Ada", password:"password123"}) { accessToken user { id email } userErrors { code } } }`
	out := postGraphQL(t, app.Handler, mutation)

	if errs, ok := out["errors"]; ok {
		t.Fatalf("unexpected errors: %v", errs)
	}

	data, _ := out["data"].(map[string]any)
	register, _ := data["register"].(map[string]any)
	if register == nil {
		t.Fatalf("missing register payload: %v", out)
	}

	token, _ := register["accessToken"].(string)
	if token == "" {
		t.Fatalf("expected non-empty access token; payload=%v", register)
	}

	user, _ := register["user"].(map[string]any)
	if user == nil || user["email"] != "a@b.com" {
		t.Fatalf("unexpected user payload: %v", register)
	}
}

func TestApp_MeRequiresAuth(t *testing.T) {
	app := newTestApp(t)

	out := postGraphQL(t, app.Handler, `query { me { id } }`)

	errs, ok := out["errors"].([]any)
	if !ok || len(errs) == 0 {
		t.Fatalf("expected errors for unauthenticated me query: %v", out)
	}

	first, _ := errs[0].(map[string]any)
	ext, _ := first["extensions"].(map[string]any)
	if ext["code"] != "UNAUTHENTICATED" {
		t.Fatalf("expected UNAUTHENTICATED, got %v", ext["code"])
	}
}
