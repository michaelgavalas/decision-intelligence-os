// Package directives implements the runtime behavior of the GraphQL schema
// directives. gqlgen generates the wiring that invokes these functions; this
// package supplies what they do.
package directives

import (
	"context"

	"github.com/99designs/gqlgen/graphql"

	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/authctx"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/errors"
)

// Authenticated enforces the @authenticated directive: it rejects the field
// resolution with an unauthenticated error unless the request context carries a
// principal, in which case it delegates to the next resolver.
func Authenticated(ctx context.Context, obj any, next graphql.Resolver) (any, error) {
	if _, ok := authctx.From(ctx); !ok {
		return nil, errors.Unauthenticated("UNAUTHENTICATED", "authentication required")
	}
	return next(ctx)
}
