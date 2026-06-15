// Package graph wires the GraphQL execution layer. This file maps the backend's
// typed errors onto GraphQL errors so clients receive a stable code and a safe
// message regardless of the underlying failure.
package graph

import (
	"context"
	stderrors "errors"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/gqlerror"

	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/errors"
)

// internalMessage is the public message returned for internal failures, so the
// underlying cause is never leaked to clients.
const internalMessage = "internal server error"

// codeFor maps a typed error Kind to the stable extension code surfaced to
// clients.
func codeFor(kind errors.Kind) string {
	switch kind {
	case errors.KindUnauthenticated:
		return "UNAUTHENTICATED"
	case errors.KindForbidden:
		return "FORBIDDEN"
	case errors.KindNotFound:
		return "NOT_FOUND"
	case errors.KindConflict:
		return "CONFLICT"
	case errors.KindValidation:
		return "VALIDATION"
	case errors.KindInternal:
		return "INTERNAL"
	default:
		return "INTERNAL"
	}
}

// messageFor returns the client-facing message for err. Internal failures get a
// generic message so their cause is never exposed; classified errors surface
// their human-readable domain message.
func messageFor(err error, kind errors.Kind) string {
	if kind == errors.KindInternal {
		return internalMessage
	}
	var typed *errors.Error
	if stderrors.As(err, &typed) {
		return typed.Message
	}
	return err.Error()
}

// PresentError converts an error into a GraphQL error suitable for gqlgen's
// SetErrorPresenter. It attaches the error's Kind as an extension "code" and,
// when present, the domain Code as an extension "reason". Internal errors are
// given a generic public message so their cause is never exposed.
func PresentError(ctx context.Context, err error) *gqlerror.Error {
	gqlErr := graphql.DefaultErrorPresenter(ctx, err)

	kind := errors.KindOf(err)
	gqlErr.Message = messageFor(err, kind)

	if gqlErr.Extensions == nil {
		gqlErr.Extensions = map[string]any{}
	}
	gqlErr.Extensions["code"] = codeFor(kind)
	if reason := errors.CodeOf(err); reason != "" {
		gqlErr.Extensions["reason"] = reason
	}

	return gqlErr
}
