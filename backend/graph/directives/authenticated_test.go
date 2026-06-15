package directives_test

import (
	"context"
	"testing"

	"github.com/michaelgavalas/decision-intelligence-os/backend/graph/directives"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/authctx"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/errors"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/id"
)

func TestAuthenticatedRejectsAnonymous(t *testing.T) {
	called := false
	next := func(context.Context) (any, error) {
		called = true
		return "ok", nil
	}

	res, err := directives.Authenticated(context.Background(), nil, next)
	if err == nil {
		t.Fatal("Authenticated(anonymous) returned nil error, want error")
	}
	if errors.KindOf(err) != errors.KindUnauthenticated {
		t.Errorf("KindOf(err) = %v, want KindUnauthenticated", errors.KindOf(err))
	}
	if called {
		t.Error("next resolver was called for an anonymous request")
	}
	if res != nil {
		t.Errorf("result = %v, want nil", res)
	}
}

func TestAuthenticatedAllowsPrincipal(t *testing.T) {
	called := false
	next := func(context.Context) (any, error) {
		called = true
		return "ok", nil
	}

	ctx := authctx.WithPrincipal(context.Background(), authctx.Principal{
		UserID: id.New(),
		Role:   "member",
	})

	res, err := directives.Authenticated(ctx, nil, next)
	if err != nil {
		t.Fatalf("Authenticated(principal) returned error: %v", err)
	}
	if !called {
		t.Error("next resolver was not called for an authenticated request")
	}
	if res != "ok" {
		t.Errorf("result = %v, want %q", res, "ok")
	}
}
