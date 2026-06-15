package auth

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/authctx"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/errors"
)

// refreshTokenBytes is the length of the random secret behind an opaque refresh
// token. 32 bytes (256 bits) is well beyond brute-force reach.
const refreshTokenBytes = 32

// accessClaims is the JWT payload for an access token. Standard registered
// claims carry the subject (user id) and expiry; the team and role are custom
// claims that mirror the request Principal.
type accessClaims struct {
	jwt.RegisteredClaims
	TeamID string `json:"team_id,omitempty"`
	Role   string `json:"role,omitempty"`
}

// GenerateKeyPair returns a fresh ed25519 key pair. It is intended for
// development and tests; production keys are loaded from configuration.
func GenerateKeyPair() (ed25519.PublicKey, ed25519.PrivateKey, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, errors.Wrap(err, errors.KindInternal, "KEYPAIR_FAILED", "failed to generate key pair")
	}
	return pub, priv, nil
}

// IssueAccessToken signs an EdDSA JWT for the principal, with issued-at set to
// now and expiry at now+ttl.
func IssueAccessToken(priv ed25519.PrivateKey, p authctx.Principal, now time.Time, ttl time.Duration) (string, error) {
	claims := accessClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   p.UserID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
		Role: p.Role,
	}
	if p.TeamID != nil {
		claims.TeamID = p.TeamID.String()
	}

	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	signed, err := token.SignedString(priv)
	if err != nil {
		return "", errors.Wrap(err, errors.KindInternal, "TOKEN_SIGN_FAILED", "failed to sign access token")
	}
	return signed, nil
}

// parseOptions configures ParseAccessToken.
type parseOptions struct {
	timeFunc func() time.Time
}

// ParseOption customizes token parsing.
type ParseOption func(*parseOptions)

// WithClock makes expiry validation use now() as its time source instead of the
// real wall clock. Production passes the system clock; tests pass a fixed clock so
// issuance and validation share one time source.
func WithClock(now func() time.Time) ParseOption {
	return func(o *parseOptions) { o.timeFunc = now }
}

// ParseAccessToken validates an access token's signature and expiry against pub
// and returns the embedded Principal. Any failure maps to an unauthenticated
// error so the transport layer can respond uniformly.
func ParseAccessToken(pub ed25519.PublicKey, token string, opts ...ParseOption) (authctx.Principal, error) {
	var options parseOptions
	for _, opt := range opts {
		opt(&options)
	}

	parserOpts := []jwt.ParserOption{jwt.WithValidMethods([]string{"EdDSA"})}
	if options.timeFunc != nil {
		parserOpts = append(parserOpts, jwt.WithTimeFunc(options.timeFunc))
	}

	var claims accessClaims
	parsed, err := jwt.ParseWithClaims(token, &claims, func(_ *jwt.Token) (any, error) {
		return pub, nil
	}, parserOpts...)
	if err != nil || !parsed.Valid {
		return authctx.Principal{}, errors.Unauthenticated("INVALID_TOKEN", "invalid or expired access token")
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return authctx.Principal{}, errors.Unauthenticated("INVALID_TOKEN", "invalid or expired access token")
	}

	principal := authctx.Principal{UserID: userID, Role: claims.Role}
	if claims.TeamID != "" {
		teamID, err := uuid.Parse(claims.TeamID)
		if err != nil {
			return authctx.Principal{}, errors.Unauthenticated("INVALID_TOKEN", "invalid or expired access token")
		}
		principal.TeamID = &teamID
	}
	return principal, nil
}

// GenerateRefreshToken returns a new opaque refresh token: a base64url-encoded
// random secret given to the client, and its SHA-256 hex digest for storage.
// Only the hash is persisted, so a database leak cannot reveal usable tokens.
func GenerateRefreshToken() (raw string, hash string, err error) {
	buf := make([]byte, refreshTokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", "", errors.Wrap(err, errors.KindInternal, "REFRESH_GEN_FAILED", "failed to generate refresh token")
	}
	raw = base64.RawURLEncoding.EncodeToString(buf)
	return raw, HashRefreshToken(raw), nil
}

// HashRefreshToken returns the SHA-256 hex digest of a raw refresh token.
func HashRefreshToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
