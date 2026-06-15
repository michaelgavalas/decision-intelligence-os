package config

import (
	"testing"
	"time"

	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/errors"
)

// envFunc builds a getenv function backed by a map.
func envFunc(m map[string]string) func(string) string {
	return func(key string) string {
		return m[key]
	}
}

func TestLoadDefaults(t *testing.T) {
	cfg, err := Load(envFunc(map[string]string{
		"DATABASE_URL": "postgres://localhost/app",
	}))
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	checks := []struct {
		name string
		got  any
		want any
	}{
		{"Env", cfg.Env, "development"},
		{"Port", cfg.Port, "8080"},
		{"LogLevel", cfg.LogLevel, "info"},
		{"AccessTokenTTL", cfg.AccessTokenTTL, 15 * time.Minute},
		{"RefreshTokenTTL", cfg.RefreshTokenTTL, 720 * time.Hour},
		{"CookieDomain", cfg.CookieDomain, "localhost"},
		{"CookieSecure", cfg.CookieSecure, false},
		{"GraphQLIntrospection", cfg.GraphQLIntrospection, true},
		{"GraphQLComplexityLimit", cfg.GraphQLComplexityLimit, 300},
		{"AIEnabled", cfg.AIEnabled, false},
		{"AIModel", cfg.AIModel, "claude-opus-4-8"},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s = %v, want %v", c.name, c.got, c.want)
		}
	}

	if len(cfg.CORSAllowedOrigins) != 1 || cfg.CORSAllowedOrigins[0] != "http://localhost:5173" {
		t.Errorf("CORSAllowedOrigins = %v, want [http://localhost:5173]", cfg.CORSAllowedOrigins)
	}
	if cfg.IsProduction() {
		t.Errorf("IsProduction() = true, want false for development env")
	}
}

func TestLoadMissingDatabaseURL(t *testing.T) {
	_, err := Load(envFunc(map[string]string{}))
	if err == nil {
		t.Fatal("expected error for missing DATABASE_URL, got nil")
	}
	if errors.KindOf(err) != errors.KindValidation {
		t.Errorf("KindOf = %v, want KindValidation", errors.KindOf(err))
	}
}

func TestLoadProductionRequiresSecrets(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
	}{
		{
			name: "missing jwt private key",
			env: map[string]string{
				"ENV":            "production",
				"DATABASE_URL":   "postgres://db/app",
				"JWT_PUBLIC_KEY": "pub",
				"CSRF_SECRET":    "secret",
			},
		},
		{
			name: "missing jwt public key",
			env: map[string]string{
				"ENV":             "production",
				"DATABASE_URL":    "postgres://db/app",
				"JWT_PRIVATE_KEY": "priv",
				"CSRF_SECRET":     "secret",
			},
		},
		{
			name: "missing csrf secret",
			env: map[string]string{
				"ENV":             "production",
				"DATABASE_URL":    "postgres://db/app",
				"JWT_PRIVATE_KEY": "priv",
				"JWT_PUBLIC_KEY":  "pub",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Load(envFunc(tt.env))
			if err == nil {
				t.Fatal("expected validation error, got nil")
			}
			if errors.KindOf(err) != errors.KindValidation {
				t.Errorf("KindOf = %v, want KindValidation", errors.KindOf(err))
			}
		})
	}
}

func TestLoadProductionComplete(t *testing.T) {
	cfg, err := Load(envFunc(map[string]string{
		"ENV":             "production",
		"DATABASE_URL":    "postgres://db/app",
		"JWT_PRIVATE_KEY": "priv",
		"JWT_PUBLIC_KEY":  "pub",
		"CSRF_SECRET":     "secret",
	}))
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if !cfg.IsProduction() {
		t.Error("IsProduction() = false, want true")
	}
}

func TestLoadCORSSplitAndTrim(t *testing.T) {
	cfg, err := Load(envFunc(map[string]string{
		"DATABASE_URL":         "postgres://db/app",
		"CORS_ALLOWED_ORIGINS": "http://a.com, http://b.com ,http://c.com",
	}))
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	want := []string{"http://a.com", "http://b.com", "http://c.com"}
	if len(cfg.CORSAllowedOrigins) != len(want) {
		t.Fatalf("CORSAllowedOrigins = %v, want %v", cfg.CORSAllowedOrigins, want)
	}
	for i := range want {
		if cfg.CORSAllowedOrigins[i] != want[i] {
			t.Errorf("origin[%d] = %q, want %q", i, cfg.CORSAllowedOrigins[i], want[i])
		}
	}
}

func TestLoadBadDurationFallsBack(t *testing.T) {
	cfg, err := Load(envFunc(map[string]string{
		"DATABASE_URL":     "postgres://db/app",
		"ACCESS_TOKEN_TTL": "not-a-duration",
	}))
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.AccessTokenTTL != 15*time.Minute {
		t.Errorf("AccessTokenTTL = %v, want 15m fallback", cfg.AccessTokenTTL)
	}
}

func TestLoadValidDurationParsed(t *testing.T) {
	cfg, err := Load(envFunc(map[string]string{
		"DATABASE_URL":      "postgres://db/app",
		"ACCESS_TOKEN_TTL":  "30m",
		"REFRESH_TOKEN_TTL": "168h",
	}))
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.AccessTokenTTL != 30*time.Minute {
		t.Errorf("AccessTokenTTL = %v, want 30m", cfg.AccessTokenTTL)
	}
	if cfg.RefreshTokenTTL != 168*time.Hour {
		t.Errorf("RefreshTokenTTL = %v, want 168h", cfg.RefreshTokenTTL)
	}
}

func TestLoadBoolParsing(t *testing.T) {
	tests := []struct {
		value string
		want  bool
	}{
		{"true", true},
		{"1", true},
		{"false", false},
		{"0", false},
		{"garbage", false}, // falls back to documented default (false)
	}
	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			cfg, err := Load(envFunc(map[string]string{
				"DATABASE_URL":  "postgres://db/app",
				"COOKIE_SECURE": tt.value,
			}))
			if err != nil {
				t.Fatalf("Load returned error: %v", err)
			}
			if cfg.CookieSecure != tt.want {
				t.Errorf("CookieSecure = %v, want %v", cfg.CookieSecure, tt.want)
			}
		})
	}
}

func TestLoadGraphQLComplexityOverride(t *testing.T) {
	cfg, err := Load(envFunc(map[string]string{
		"DATABASE_URL":             "postgres://db/app",
		"GRAPHQL_COMPLEXITY_LIMIT": "500",
	}))
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.GraphQLComplexityLimit != 500 {
		t.Errorf("GraphQLComplexityLimit = %d, want 500", cfg.GraphQLComplexityLimit)
	}
}

func TestLoadBadComplexityFallsBack(t *testing.T) {
	cfg, err := Load(envFunc(map[string]string{
		"DATABASE_URL":             "postgres://db/app",
		"GRAPHQL_COMPLEXITY_LIMIT": "abc",
	}))
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.GraphQLComplexityLimit != 300 {
		t.Errorf("GraphQLComplexityLimit = %d, want 300 fallback", cfg.GraphQLComplexityLimit)
	}
}
