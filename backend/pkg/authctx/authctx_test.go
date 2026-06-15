package authctx_test

import (
	"context"
	"testing"

	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/authctx"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/errors"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/id"
)

func TestRoundTrip(t *testing.T) {
	team := id.New()
	want := authctx.Principal{
		UserID: id.New(),
		TeamID: &team,
		Role:   "admin",
	}

	ctx := authctx.WithPrincipal(context.Background(), want)
	got, ok := authctx.From(ctx)
	if !ok {
		t.Fatal("From returned ok=false after WithPrincipal")
	}
	if got.UserID != want.UserID || got.Role != want.Role {
		t.Errorf("From = %+v, want %+v", got, want)
	}
	if got.TeamID == nil || *got.TeamID != team {
		t.Errorf("TeamID = %v, want %v", got.TeamID, team)
	}
}

func TestFromEmpty(t *testing.T) {
	if _, ok := authctx.From(context.Background()); ok {
		t.Error("From(empty) returned ok=true, want false")
	}
}

func TestRequire(t *testing.T) {
	p := authctx.Principal{UserID: id.New(), Role: "member"}
	ctx := authctx.WithPrincipal(context.Background(), p)
	got, err := authctx.Require(ctx)
	if err != nil {
		t.Fatalf("Require returned error: %v", err)
	}
	if got.UserID != p.UserID {
		t.Errorf("Require = %+v, want %+v", got, p)
	}
}

func TestRequireEmpty(t *testing.T) {
	_, err := authctx.Require(context.Background())
	if err == nil {
		t.Fatal("Require(empty) returned nil error, want error")
	}
	if errors.KindOf(err) != errors.KindUnauthenticated {
		t.Errorf("KindOf(err) = %v, want KindUnauthenticated", errors.KindOf(err))
	}
}
