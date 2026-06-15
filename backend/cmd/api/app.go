package main

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"

	"github.com/michaelgavalas/decision-intelligence-os/backend/graph"
	"github.com/michaelgavalas/decision-intelligence-os/backend/graph/directives"
	"github.com/michaelgavalas/decision-intelligence-os/backend/graph/loaders"
	"github.com/michaelgavalas/decision-intelligence-os/backend/graph/resolvers"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/ai"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/analytics"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/assumptions"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/auth"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/decisions"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/evidence"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/outcomes"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/config"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/db"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/events"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/health"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/httpx"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/pubsub"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/ratelimit"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/predictions"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/teams"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/users"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/authctx"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/clock"
)

const (
	// queryCacheSize bounds the parsed-query cache.
	queryCacheSize = 1000
	// wsKeepAlive is the interval between subscription keep-alive pings.
	wsKeepAlive = 10 * time.Second
)

// App holds the constructed HTTP handler and the resources that must be closed
// on shutdown.
type App struct {
	Handler http.Handler
	pool    *pgxpool.Pool
}

// Close releases resources held by the App.
func (a *App) Close() {
	if a.pool != nil {
		a.pool.Close()
	}
}

// NewApp builds the entire object graph from cfg: the connection pool,
// transaction manager, event recorder, pub/sub, rate limiter, repositories,
// services, AI provider, GraphQL resolver and server, and the HTTP router. It
// does not run migrations; those run as a separate deploy step.
func NewApp(ctx context.Context, cfg config.Config, log *slog.Logger) (*App, error) {
	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	txm := db.NewTxManager(pool)
	recorder := events.NewRecorder()
	pub := pubsub.NewPublisher(pool)
	sub := pubsub.NewSubscriber(pool)
	limiter := ratelimit.NewLimiter(pool)
	clk := clock.System{}

	// Repositories are stateless; each operation receives a Querier.
	usersRepo := users.NewRepository()
	teamsRepo := teams.NewRepository()
	decisionsRepo := decisions.NewRepository()
	assumptionsRepo := assumptions.NewRepository()
	evidenceRepo := evidence.NewRepository()
	predictionsRepo := predictions.NewRepository()
	outcomesRepo := outcomes.NewRepository()
	analyticsRepo := analytics.NewRepository()
	authRepo := auth.NewRepository()

	usersSvc := users.NewService(pool, txm, usersRepo, clk)
	teamsSvc := teams.NewService(pool, txm, teamsRepo, clk)

	tokenPub, priv, err := loadTokenKeys(cfg, log)
	if err != nil {
		pool.Close()
		return nil, err
	}
	tokenCfg := auth.TokenConfig{
		PrivateKey: priv,
		PublicKey:  tokenPub,
		AccessTTL:  cfg.AccessTokenTTL,
		RefreshTTL: cfg.RefreshTokenTTL,
	}
	authSvc := auth.NewService(pool, txm, usersSvc, teamsSvc, authRepo, limiter, tokenCfg, clk)

	decisionsSvc := decisions.NewService(pool, txm, decisionsRepo, recorder, teamsSvc, clk)
	assumptionsSvc := assumptions.NewService(pool, txm, assumptionsRepo, recorder, decisionsSvc)
	evidenceSvc := evidence.NewService(pool, txm, evidenceRepo, recorder, assumptionsSvc)
	predictionsSvc := predictions.NewService(pool, txm, predictionsRepo, recorder, decisionsSvc)
	outcomesSvc := outcomes.NewService(pool, txm, outcomesRepo, recorder, decisionsSvc, clk)
	analyticsSvc := analytics.NewService(pool, analyticsRepo, teamsSvc)

	var provider ai.LLMProvider = ai.Disabled{}
	if cfg.AIEnabled && cfg.AnthropicAPIKey != "" {
		provider = ai.NewClaudeProvider(cfg.AnthropicAPIKey, cfg.AIModel, "", nil)
	}
	aiSvc := ai.NewService(cfg.AIEnabled, provider)

	resolver := resolvers.NewResolver(
		authSvc,
		usersSvc,
		teamsSvc,
		decisionsSvc,
		assumptionsSvc,
		evidenceSvc,
		predictionsSvc,
		outcomesSvc,
		analyticsSvc,
		aiSvc,
		pub,
		sub,
		resolvers.CookieOptions{
			Secure:     cfg.CookieSecure,
			Domain:     cfg.CookieDomain,
			RefreshTTL: cfg.RefreshTokenTTL,
		},
	)

	srv := newGraphQLServer(cfg, resolver, authSvc)

	r := chi.NewRouter()
	r.Use(httpx.RequestIDMiddleware)
	r.Use(httpx.LoggerMiddleware(log))
	r.Use(middleware.Recoverer)
	r.Use(securityHeaders(cfg))
	r.Use(corsMiddleware(cfg))
	r.Use(httpx.ClientIPMiddleware)
	r.Use(httpx.HTTPContext)
	r.Use(authExtractor(authSvc))
	r.Use(loaders.Middleware(pool, usersRepo, teamsRepo, decisionsRepo, assumptionsRepo, evidenceRepo, predictionsRepo, outcomesRepo))

	r.Get("/healthz", health.Liveness())
	r.Get("/readyz", health.Readiness(pool))
	r.Handle("/graphql", srv)

	return &App{Handler: r, pool: pool}, nil
}

// newGraphQLServer constructs the gqlgen server with the configured transports,
// query cache, complexity limit, error presenter, and panic recovery.
func newGraphQLServer(cfg config.Config, resolver *resolvers.Resolver, authSvc auth.Service) *handler.Server {
	es := graph.NewExecutableSchema(graph.Config{
		Resolvers:  resolver,
		Directives: graph.DirectiveRoot{Authenticated: directives.Authenticated},
	})

	srv := handler.New(es)
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.Websocket{
		KeepAlivePingInterval: wsKeepAlive,
		InitFunc:              wsInitFunc(authSvc),
	})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](queryCacheSize))

	if cfg.GraphQLIntrospection {
		srv.Use(extension.Introspection{})
	}
	srv.Use(extension.FixedComplexityLimit(cfg.GraphQLComplexityLimit))

	srv.SetErrorPresenter(graph.PresentError)
	srv.SetRecoverFunc(func(ctx context.Context, err any) error {
		httpx.Logger(ctx).Error("graphql panic", "panic", err)
		return gqlerror.Errorf("internal server error")
	})

	return srv
}

// wsInitFunc authenticates GraphQL subscriptions from the websocket connection
// init payload. It reads the optional "authorization" field, strips a "Bearer "
// prefix, and on a valid token attaches the principal to the context. Missing or
// invalid tokens proceed unauthenticated; the @authenticated directive then
// rejects operations that require a principal.
func wsInitFunc(authSvc auth.Service) transport.WebsocketInitFunc {
	return func(ctx context.Context, initPayload transport.InitPayload) (context.Context, *transport.InitPayload, error) {
		raw, _ := initPayload["authorization"].(string)
		token := strings.TrimSpace(strings.TrimPrefix(raw, "Bearer "))
		if token == "" {
			return ctx, nil, nil
		}
		principal, err := authSvc.ParseAccessToken(token)
		if err != nil {
			// An invalid token connects unauthenticated; the @authenticated
			// directive rejects operations that require a principal.
			return ctx, nil, nil //nolint:nilerr // intentional: proceed unauthenticated
		}
		return authctx.WithPrincipal(ctx, principal), nil, nil
	}
}

// securityHeaders sets conservative security response headers on every request.
// HSTS is only emitted when cfg.CookieSecure is set, since advertising it over
// plain HTTP (as in local development) would wrongly pin clients to HTTPS.
func securityHeaders(cfg config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := w.Header()
			h.Set("X-Content-Type-Options", "nosniff")
			h.Set("X-Frame-Options", "DENY")
			h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
			h.Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
			h.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
			if cfg.CookieSecure {
				h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			}
			next.ServeHTTP(w, r)
		})
	}
}

// corsMiddleware builds the CORS handler from configuration.
func corsMiddleware(cfg config.Config) func(http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORSAllowedOrigins,
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodOptions},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	})
}

// authExtractor reads a Bearer access token from the Authorization header and,
// when valid, attaches the principal to the request context. It never fails the
// request on a missing or invalid token; authorization is enforced downstream.
func authExtractor(authSvc auth.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header != "" {
				token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
				if token != "" {
					if principal, err := authSvc.ParseAccessToken(token); err == nil {
						r = r.WithContext(authctx.WithPrincipal(r.Context(), principal))
					}
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
