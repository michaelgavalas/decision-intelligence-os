// Package health exposes liveness and readiness HTTP handlers. Liveness
// reflects process health; readiness reflects the ability to serve traffic,
// which depends on a reachable database.
package health

import (
	"context"
	"encoding/json"
	"net/http"
)

// Pinger reports whether a dependency is reachable. It is satisfied by
// *pgxpool.Pool.
type Pinger interface {
	Ping(ctx context.Context) error
}

// statusResponse is the JSON body returned by the health handlers.
type statusResponse struct {
	Status string `json:"status"`
}

// writeStatus writes the given HTTP status code with a JSON status body.
func writeStatus(w http.ResponseWriter, code int, status string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(statusResponse{Status: status})
}

// Liveness reports that the process is running. It always returns 200.
func Liveness() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		writeStatus(w, http.StatusOK, "ok")
	}
}

// Readiness reports whether the service can serve traffic by pinging the
// dependency. It returns 200 on success and 503 when the ping fails.
func Readiness(p Pinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := p.Ping(r.Context()); err != nil {
			writeStatus(w, http.StatusServiceUnavailable, "unavailable")
			return
		}
		writeStatus(w, http.StatusOK, "ok")
	}
}
