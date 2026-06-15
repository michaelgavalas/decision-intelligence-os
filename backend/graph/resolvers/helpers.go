package resolvers

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"

	"github.com/michaelgavalas/decision-intelligence-os/backend/graph/model"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/auth"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/httpx"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/errors"
)

// decisionUpdatedChannel is the pubsub channel real-time decision changes are
// published on and subscribed from.
const decisionUpdatedChannel = "decision_updated"

// decisionSubscriptionBuffer bounds the per-subscriber delivery channel so a
// slow client cannot stall the fan-out goroutine indefinitely.
const decisionSubscriptionBuffer = 8

// decisionEvent is the JSON payload published when a decision changes, carrying
// just the ids a subscriber needs to filter and reload the decision.
type decisionEvent struct {
	TeamID     string `json:"team_id"`
	DecisionID string `json:"decision_id"`
}

// toUserErrors maps an expected, recoverable error (validation or conflict)
// onto the userErrors payload pattern. It returns handled=true only for those
// kinds; any other error (forbidden, unauthenticated, not found, internal) is
// left for the caller to surface as a transport error so the client still sees
// the right GraphQL error code.
func toUserErrors(err error) ([]model.UserError, bool) {
	switch errors.KindOf(err) {
	case errors.KindValidation, errors.KindConflict:
		return []model.UserError{{
			Message: messageOf(err),
			Code:    errors.CodeOf(err),
		}}, true
	default:
		return nil, false
	}
}

// noUserErrors is the empty (non-nil) slice returned with a successful mutation
// payload, since the schema types userErrors as a non-null list.
func noUserErrors() []model.UserError {
	return []model.UserError{}
}

// messageOf returns the human-readable message of a typed error, falling back
// to the error string for untyped errors.
func messageOf(err error) string {
	var typed *errors.Error
	if asTyped(err, &typed) {
		return typed.Message
	}
	return err.Error()
}

// asTyped is a thin wrapper over errors.As kept local so this file does not
// import the standard errors package alongside the project's errors package.
func asTyped(err error, target **errors.Error) bool {
	for err != nil {
		if e, ok := err.(*errors.Error); ok { //nolint:errorlint // direct match is sufficient; Unwrap is walked below
			*target = e
			return true
		}
		u, ok := err.(interface{ Unwrap() error })
		if !ok {
			return false
		}
		err = u.Unwrap()
	}
	return false
}

// publishDecisionUpdated emits a real-time decision-changed event. Publish
// failures are intentionally swallowed: they must never fail the mutation that
// triggered them, since the write itself already succeeded.
func (r *Resolver) publishDecisionUpdated(ctx context.Context, teamID, decisionID uuid.UUID) {
	if r.Publisher == nil {
		return
	}
	payload, err := json.Marshal(decisionEvent{
		TeamID:     teamID.String(),
		DecisionID: decisionID.String(),
	})
	if err != nil {
		return
	}
	_ = r.Publisher.Publish(ctx, decisionUpdatedChannel, string(payload))
}

// authPayload maps a successful auth result onto the GraphQL payload. The raw
// refresh token is returned to the caller; the transport layer decides whether
// to also set it as an httpOnly cookie.
func authPayload(result auth.AuthResult) *model.AuthPayload {
	accessToken := result.AccessToken
	accessExpiresAt := result.AccessExpiresAt
	refreshToken := result.RefreshToken
	return &model.AuthPayload{
		User:            toUser(result.User),
		AccessToken:     &accessToken,
		AccessExpiresAt: &accessExpiresAt,
		RefreshToken:    &refreshToken,
		UserErrors:      noUserErrors(),
	}
}

// rateKey returns the login rate-limit key for the request: the client IP so
// throttling is per-source. It falls back to "unknown" when the IP is not
// available, which lumps such requests under a single shared bucket.
func rateKey(ctx context.Context) string {
	if ip := httpx.ClientIP(ctx); ip != "" {
		return ip
	}
	return "unknown"
}
