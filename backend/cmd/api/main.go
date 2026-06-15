// Command api is the Decision Intelligence OS backend: a single HTTP service
// exposing the GraphQL API and health endpoints. It is the composition root that
// builds the object graph and runs the server.
package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/config"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/logging"
)

const (
	// readHeaderTimeout bounds how long the server waits to read request headers,
	// mitigating slow-client attacks.
	readHeaderTimeout = 10 * time.Second
	// shutdownTimeout bounds graceful shutdown before connections are forced
	// closed.
	shutdownTimeout = 15 * time.Second
	// healthcheckTimeout bounds the in-process health probe used by -healthcheck.
	healthcheckTimeout = 3 * time.Second
)

func main() {
	healthcheck := flag.Bool("healthcheck", false, "probe the local health endpoint and exit (used by the container healthcheck)")
	flag.Parse()

	if *healthcheck {
		os.Exit(runHealthcheck())
	}

	if err := run(); err != nil {
		// run already logs the failure; signal a non-zero exit to the supervisor.
		os.Exit(1)
	}
}

// run loads configuration, builds the App, and serves until a termination
// signal triggers graceful shutdown.
func run() error {
	cfg, err := config.Load(os.Getenv)
	if err != nil {
		// The logger needs config for its level, but a config failure must still be
		// visible; use a default logger.
		logging.New(os.Stdout, "info").Error("failed to load configuration", "error", err)
		return err
	}

	log := logging.New(os.Stdout, cfg.LogLevel)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	app, err := NewApp(ctx, cfg, log)
	if err != nil {
		log.Error("failed to build application", "error", err)
		return err
	}
	defer app.Close()

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           app.Handler,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	serverErr := make(chan error, 1)
	go func() {
		log.Info("server starting", "port", cfg.Port, "env", cfg.Env)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		log.Error("server failed", "error", err)
		return err
	case <-ctx.Done():
		log.Info("shutdown signal received; draining connections")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("graceful shutdown failed", "error", err)
		return err
	}

	log.Info("server stopped")
	return nil
}

// runHealthcheck issues a GET to the local liveness endpoint and returns a
// process exit code: 0 when it responds 200, 1 otherwise. It backs the container
// healthcheck so no extra tooling is needed in the image.
func runHealthcheck() int {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	ctx, cancel := context.WithTimeout(context.Background(), healthcheckTimeout)
	defer cancel()

	// Self-probe against the loopback interface for the container healthcheck; the only
	// variable is the local listen port, never a user-supplied address.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://127.0.0.1:"+port+"/healthz", nil)
	if err != nil {
		return 1
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 1
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return 1
	}
	return 0
}
