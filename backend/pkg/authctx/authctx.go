// Package authctx carries the authenticated Principal through the request
// context.
package authctx

import (
	"context"

	"github.com/google/uuid"

	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/errors"
)

// Principal identifies the authenticated caller and the team scope under which
// the current request operates.
type Principal struct {
	UserID uuid.UUID
	TeamID *uuid.UUID
	Role   string
}

type contextKey struct{}

var principalKey = contextKey{}

// WithPrincipal returns a copy of ctx that carries the given principal.
func WithPrincipal(ctx context.Context, p Principal) context.Context {
	return context.WithValue(ctx, principalKey, p)
}

// From returns the principal stored in ctx, if any.
func From(ctx context.Context) (Principal, bool) {
	p, ok := ctx.Value(principalKey).(Principal)
	return p, ok
}

// Require returns the principal stored in ctx, or an unauthenticated error when
// none is present.
func Require(ctx context.Context) (Principal, error) {
	p, ok := From(ctx)
	if !ok {
		return Principal{}, errors.Unauthenticated("UNAUTHENTICATED", "authentication required")
	}
	return p, nil
}
