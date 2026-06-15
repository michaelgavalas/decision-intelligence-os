package auth

import (
	"time"

	"github.com/google/uuid"

	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/users"
)

// RefreshToken is a persisted, hashed refresh token. The raw token is never
// stored; only TokenHash is. A token is usable when it is neither revoked nor
// expired.
type RefreshToken struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	TokenHash  string
	ExpiresAt  time.Time
	RevokedAt  *time.Time
	ReplacedBy *uuid.UUID
	CreatedAt  time.Time
}

// Revoked reports whether the token has been revoked.
func (t RefreshToken) Revoked() bool {
	return t.RevokedAt != nil
}

// Expired reports whether the token's lifetime has elapsed as of now.
func (t RefreshToken) Expired(now time.Time) bool {
	return !now.Before(t.ExpiresAt)
}

// AuthResult is the outcome of a successful authentication flow. The transport
// layer returns the access token to the client and stores the raw refresh token
// in an httpOnly cookie.
type AuthResult struct { //nolint:revive // AuthResult is the established cross-domain name for an auth flow's output.
	User             users.User
	AccessToken      string
	AccessExpiresAt  time.Time
	RefreshToken     string
	RefreshExpiresAt time.Time
}

// RegisterInput carries the fields needed to register a new account.
type RegisterInput struct {
	Email    string
	Name     string
	Password string
}
