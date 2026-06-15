// Package config loads and validates application configuration from environment
// variables. Loading is hermetic: callers pass a getenv function so the same
// logic is exercised in tests without touching the process environment.
package config

import (
	"strconv"
	"strings"
	"time"

	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/errors"
)

// Config holds the fully resolved application configuration.
type Config struct {
	Env                    string        // "development" | "production"
	Port                   string        // default "8080"
	LogLevel               string        // default "info"
	DatabaseURL            string        // required
	JWTPrivateKey          string        // required in production
	JWTPublicKey           string        // required in production
	AccessTokenTTL         time.Duration // default 15m
	RefreshTokenTTL        time.Duration // default 720h
	CookieDomain           string        // default "localhost"
	CookieSecure           bool          // default false
	CSRFSecret             string        // required in production
	CORSAllowedOrigins     []string      // comma-separated; default ["http://localhost:5173"]
	GraphQLIntrospection   bool          // default true
	GraphQLComplexityLimit int           // default 300
	AIEnabled              bool          // default false
	AIModel                string        // default "claude-opus-4-8"
	AnthropicAPIKey        string        // optional
}

// Default values applied when an environment variable is unset or invalid.
const (
	defaultEnv             = "development"
	defaultPort            = "8080"
	defaultLogLevel        = "info"
	defaultAccessTokenTTL  = 15 * time.Minute
	defaultRefreshTokenTTL = 720 * time.Hour
	defaultCookieDomain    = "localhost"
	defaultComplexityLimit = 300
	defaultAIModel         = "claude-opus-4-8"
	defaultCORSOrigin      = "http://localhost:5173"
)

// IsProduction reports whether the configuration targets the production
// environment.
func (c Config) IsProduction() bool {
	return c.Env == "production"
}

// Load reads configuration from environment variables via getenv, applies
// defaults, and validates the result. It returns a KindValidation error when a
// required value is missing.
func Load(getenv func(string) string) (Config, error) {
	cfg := Config{
		Env:                    stringOr(getenv("ENV"), defaultEnv),
		Port:                   stringOr(getenv("PORT"), defaultPort),
		LogLevel:               stringOr(getenv("LOG_LEVEL"), defaultLogLevel),
		DatabaseURL:            getenv("DATABASE_URL"),
		JWTPrivateKey:          getenv("JWT_PRIVATE_KEY"),
		JWTPublicKey:           getenv("JWT_PUBLIC_KEY"),
		AccessTokenTTL:         durationOr(getenv("ACCESS_TOKEN_TTL"), defaultAccessTokenTTL),
		RefreshTokenTTL:        durationOr(getenv("REFRESH_TOKEN_TTL"), defaultRefreshTokenTTL),
		CookieDomain:           stringOr(getenv("COOKIE_DOMAIN"), defaultCookieDomain),
		CookieSecure:           boolOr(getenv("COOKIE_SECURE"), false),
		CSRFSecret:             getenv("CSRF_SECRET"),
		CORSAllowedOrigins:     splitOrigins(getenv("CORS_ALLOWED_ORIGINS")),
		GraphQLIntrospection:   boolOr(getenv("GRAPHQL_INTROSPECTION"), true),
		GraphQLComplexityLimit: intOr(getenv("GRAPHQL_COMPLEXITY_LIMIT"), defaultComplexityLimit),
		AIEnabled:              boolOr(getenv("AI_ENABLED"), false),
		AIModel:                stringOr(getenv("AI_MODEL"), defaultAIModel),
		AnthropicAPIKey:        getenv("ANTHROPIC_API_KEY"),
	}

	if err := cfg.validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// validate enforces required values for the resolved environment.
func (c Config) validate() error {
	var missing []string
	if c.DatabaseURL == "" {
		missing = append(missing, "DATABASE_URL")
	}
	if c.IsProduction() {
		if c.JWTPrivateKey == "" {
			missing = append(missing, "JWT_PRIVATE_KEY")
		}
		if c.JWTPublicKey == "" {
			missing = append(missing, "JWT_PUBLIC_KEY")
		}
		if c.CSRFSecret == "" {
			missing = append(missing, "CSRF_SECRET")
		}
	}
	if len(missing) > 0 {
		return errors.Validation(
			"CONFIG_MISSING_REQUIRED",
			"missing required configuration: "+strings.Join(missing, ", "),
		)
	}
	return nil
}

// stringOr returns value when non-empty, otherwise fallback.
func stringOr(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

// boolOr parses a boolean accepting "true"/"1"/"false"/"0", returning fallback
// for empty or unrecognized input.
func boolOr(value string, fallback bool) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "true", "1":
		return true
	case "false", "0":
		return false
	default:
		return fallback
	}
}

// intOr parses an integer, returning fallback for empty or invalid input.
func intOr(value string, fallback int) int {
	n, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return fallback
	}
	return n
}

// durationOr parses a Go duration, returning fallback for empty or invalid
// input.
func durationOr(value string, fallback time.Duration) time.Duration {
	d, err := time.ParseDuration(strings.TrimSpace(value))
	if err != nil {
		return fallback
	}
	return d
}

// splitOrigins splits a comma-separated origin list, trimming whitespace and
// dropping empty entries. An empty input yields the default origin.
func splitOrigins(value string) []string {
	if strings.TrimSpace(value) == "" {
		return []string{defaultCORSOrigin}
	}
	parts := strings.Split(value, ",")
	origins := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			origins = append(origins, trimmed)
		}
	}
	if len(origins) == 0 {
		return []string{defaultCORSOrigin}
	}
	return origins
}
