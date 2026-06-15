package resolvers

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

import (
	"time"

	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/ai"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/analytics"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/assumptions"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/auth"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/decisions"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/evidence"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/outcomes"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/pubsub"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/predictions"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/teams"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/users"
)

// Resolver is the root resolver and dependency-injection container for the
// GraphQL server. It holds the domain services the resolvers delegate to and
// the publish/subscribe pair that powers real-time subscriptions.
type Resolver struct {
	Auth        auth.Service
	Users       users.Service
	Teams       teams.Service
	Decisions   decisions.Service
	Assumptions assumptions.Service
	// EvidenceSvc is named with a Svc suffix because gqlgen generates an
	// Evidence() accessor method on *Resolver for the Evidence field resolver,
	// which would collide with a field named Evidence.
	EvidenceSvc evidence.Service
	Predictions predictions.Service
	Outcomes    outcomes.Service
	Analytics   analytics.Service
	AI          *ai.Service
	Publisher   *pubsub.Publisher
	Subscriber  *pubsub.Subscriber

	// CookieSecure marks auth cookies as Secure (HTTPS-only).
	CookieSecure bool
	// CookieDomain scopes auth cookies to a domain.
	CookieDomain string
	// RefreshTTL is the lifetime applied to the refresh-token cookie.
	RefreshTTL time.Duration
}

// CookieOptions carries the settings used when issuing the refresh-token and
// CSRF cookies.
type CookieOptions struct {
	Secure     bool
	Domain     string
	RefreshTTL time.Duration
}

// NewResolver wires a Resolver from its collaborators.
func NewResolver(
	a auth.Service,
	u users.Service,
	t teams.Service,
	d decisions.Service,
	as assumptions.Service,
	e evidence.Service,
	p predictions.Service,
	o outcomes.Service,
	an analytics.Service,
	aiSvc *ai.Service,
	pub *pubsub.Publisher,
	sub *pubsub.Subscriber,
	cookies CookieOptions,
) *Resolver {
	return &Resolver{
		Auth:         a,
		Users:        u,
		Teams:        t,
		Decisions:    d,
		Assumptions:  as,
		EvidenceSvc:  e,
		Predictions:  p,
		Outcomes:     o,
		Analytics:    an,
		AI:           aiSvc,
		Publisher:    pub,
		Subscriber:   sub,
		CookieSecure: cookies.Secure,
		CookieDomain: cookies.Domain,
		RefreshTTL:   cookies.RefreshTTL,
	}
}
