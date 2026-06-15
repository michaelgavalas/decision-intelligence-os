package auth

import (
	"context"
	"crypto/ed25519"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/db"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/teams"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/users"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/authctx"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/clock"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/errors"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/id"
)

// passwordMinLen is the minimum acceptable password length.
const passwordMinLen = 8

// loginRateLimit and loginRateWindow bound login attempts per key (typically
// per client IP) to slow credential-stuffing attacks.
const (
	loginRateLimit  = 10
	loginRateWindow = time.Minute
)

// UserService is the narrow slice of the users domain the auth domain depends
// on. Defining it here (rather than importing the full users.Service) keeps the
// dependency explicit and easy to fake in tests.
type UserService interface {
	Provision(ctx context.Context, q db.Querier, p users.ProvisionParams) (users.User, error)
	FindByEmail(ctx context.Context, email string) (users.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (users.User, error)
}

// TeamService is the narrow slice of the teams domain the auth domain depends
// on: provisioning a new account's personal team during registration.
type TeamService interface {
	ProvisionPersonalTeam(ctx context.Context, q db.Querier, ownerID uuid.UUID, name string) (teams.Team, error)
}

// Limiter throttles an action identified by key within a fixed window.
type Limiter interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error)
}

// txRunner runs a unit of work inside a database transaction. It is satisfied by
// *db.TxManager in production and by a fake in unit tests, so the service's
// multi-step writes can be tested without a database.
type txRunner interface {
	WithinTx(ctx context.Context, fn func(q db.Querier) error) error
}

// TokenConfig holds the signing keys and lifetimes for issued tokens.
type TokenConfig struct {
	PrivateKey ed25519.PrivateKey
	PublicKey  ed25519.PublicKey
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

// Service is the auth domain's application boundary. It owns the registration,
// login, refresh, and logout flows and exposes access-token parsing for the
// transport layer.
type Service interface {
	// Register creates a user and their personal team, then returns tokens.
	Register(ctx context.Context, in RegisterInput) (AuthResult, error)
	// Login authenticates by email and password, rate-limited per rateKey.
	Login(ctx context.Context, email, password, rateKey string) (AuthResult, error)
	// Refresh rotates a valid refresh token, returning a new token pair. A reused
	// (already-revoked) token revokes the whole family as a breach response.
	Refresh(ctx context.Context, rawRefreshToken string) (AuthResult, error)
	// Logout revokes a refresh token. It is idempotent.
	Logout(ctx context.Context, rawRefreshToken string) error
	// ParseAccessToken validates an access token and returns its Principal.
	ParseAccessToken(token string) (authctx.Principal, error)
}

// service is the default Service implementation.
type service struct {
	pool    *pgxpool.Pool
	tx      txRunner
	users   UserService
	teams   TeamService
	refresh Repository
	limiter Limiter
	cfg     TokenConfig
	clk     clock.Clock
}

// NewService wires a Service from its collaborators.
func NewService(
	pool *pgxpool.Pool,
	tx *db.TxManager,
	users UserService,
	teams TeamService,
	refresh Repository,
	limiter Limiter,
	cfg TokenConfig,
	clk clock.Clock,
) Service {
	return &service{
		pool:    pool,
		tx:      tx,
		users:   users,
		teams:   teams,
		refresh: refresh,
		limiter: limiter,
		cfg:     cfg,
		clk:     clk,
	}
}

func (s *service) Register(ctx context.Context, in RegisterInput) (AuthResult, error) {
	email := strings.TrimSpace(in.Email)
	name := strings.TrimSpace(in.Name)
	if err := validateEmail(email); err != nil {
		return AuthResult{}, err
	}
	if err := validateName(name); err != nil {
		return AuthResult{}, err
	}
	if err := validatePassword(in.Password); err != nil {
		return AuthResult{}, err
	}

	hash, err := HashPassword(in.Password)
	if err != nil {
		return AuthResult{}, err
	}

	var (
		user       users.User
		refreshRow RefreshToken
		rawRefresh string
		refreshExp = s.clk.Now().Add(s.cfg.RefreshTTL)
	)

	err = s.tx.WithinTx(ctx, func(q db.Querier) error {
		var txErr error
		user, txErr = s.users.Provision(ctx, q, users.ProvisionParams{
			Email:        email,
			Name:         name,
			PasswordHash: hash,
		})
		if txErr != nil {
			return txErr
		}
		if _, txErr = s.teams.ProvisionPersonalTeam(ctx, q, user.ID, name+"'s Team"); txErr != nil {
			return txErr
		}

		var refHash string
		rawRefresh, refHash, txErr = GenerateRefreshToken()
		if txErr != nil {
			return txErr
		}
		refreshRow, txErr = s.refresh.Store(ctx, q, RefreshToken{
			ID:        id.New(),
			UserID:    user.ID,
			TokenHash: refHash,
			ExpiresAt: refreshExp,
		})
		return txErr
	})
	if err != nil {
		return AuthResult{}, err
	}

	return s.buildResult(user, rawRefresh, refreshRow.ExpiresAt)
}

func (s *service) Login(ctx context.Context, email, password, rateKey string) (AuthResult, error) {
	allowed, err := s.limiter.Allow(ctx, "login:"+rateKey, loginRateLimit, loginRateWindow)
	if err != nil {
		return AuthResult{}, err
	}
	if !allowed {
		return AuthResult{}, errors.Unauthenticated("RATE_LIMITED", "too many login attempts, try again later")
	}

	user, err := s.users.FindByEmail(ctx, strings.TrimSpace(email))
	if err != nil {
		if errors.KindOf(err) == errors.KindNotFound {
			return AuthResult{}, errInvalidCredentials()
		}
		return AuthResult{}, err
	}

	ok, err := VerifyPassword(user.PasswordHash, password)
	if err != nil {
		return AuthResult{}, errors.Wrap(err, errors.KindInternal, "PASSWORD_VERIFY_FAILED", "failed to verify password")
	}
	if !ok {
		return AuthResult{}, errInvalidCredentials()
	}

	raw, refreshExp, err := s.issueRefresh(ctx, s.pool, user.ID)
	if err != nil {
		return AuthResult{}, err
	}
	return s.buildResult(user, raw, refreshExp)
}

func (s *service) Refresh(ctx context.Context, rawRefreshToken string) (AuthResult, error) {
	hash := HashRefreshToken(rawRefreshToken)
	token, err := s.refresh.GetByHash(ctx, s.pool, hash)
	if err != nil {
		if errors.KindOf(err) == errors.KindNotFound {
			return AuthResult{}, errors.Unauthenticated("INVALID_REFRESH", "invalid refresh token")
		}
		return AuthResult{}, err
	}

	// A revoked token presented again signals theft: the legitimate owner already
	// rotated it. Revoke the whole family and refuse.
	if token.Revoked() {
		if err := s.refresh.RevokeAllForUser(ctx, s.pool, token.UserID); err != nil {
			return AuthResult{}, err
		}
		return AuthResult{}, errors.Unauthenticated("TOKEN_REUSE", "refresh token reuse detected")
	}
	if token.Expired(s.clk.Now()) {
		return AuthResult{}, errors.Unauthenticated("EXPIRED_REFRESH", "refresh token expired")
	}

	var (
		rawRefresh string
		refreshExp time.Time
	)
	err = s.tx.WithinTx(ctx, func(q db.Querier) error {
		var txErr error
		rawRefresh, refreshExp, txErr = s.issueRefresh(ctx, q, token.UserID)
		if txErr != nil {
			return txErr
		}
		newHash := HashRefreshToken(rawRefresh)
		newRow, txErr := s.refresh.GetByHash(ctx, q, newHash)
		if txErr != nil {
			return txErr
		}
		return s.refresh.MarkReplaced(ctx, q, token.ID, newRow.ID)
	})
	if err != nil {
		return AuthResult{}, err
	}

	// users.GetByID requires an authenticated caller; the refreshing client owns
	// this token, so authorize the lookup as that user.
	authedCtx := authctx.WithPrincipal(ctx, authctx.Principal{UserID: token.UserID})
	user, err := s.users.GetByID(authedCtx, token.UserID)
	if err != nil {
		return AuthResult{}, err
	}
	return s.buildResult(user, rawRefresh, refreshExp)
}

func (s *service) Logout(ctx context.Context, rawRefreshToken string) error {
	hash := HashRefreshToken(rawRefreshToken)
	token, err := s.refresh.GetByHash(ctx, s.pool, hash)
	if err != nil {
		if errors.KindOf(err) == errors.KindNotFound {
			return nil
		}
		return err
	}
	if token.Revoked() {
		return nil
	}
	return s.refresh.Revoke(ctx, s.pool, token.ID)
}

func (s *service) ParseAccessToken(token string) (authctx.Principal, error) {
	return ParseAccessToken(s.cfg.PublicKey, token, WithClock(s.clk.Now))
}

// issueRefresh generates a refresh token and stores it against userID using the
// supplied querier, returning the raw token and its expiry.
func (s *service) issueRefresh(ctx context.Context, q db.Querier, userID uuid.UUID) (string, time.Time, error) {
	raw, hash, err := GenerateRefreshToken()
	if err != nil {
		return "", time.Time{}, err
	}
	expiresAt := s.clk.Now().Add(s.cfg.RefreshTTL)
	if _, err := s.refresh.Store(ctx, q, RefreshToken{
		ID:        id.New(),
		UserID:    userID,
		TokenHash: hash,
		ExpiresAt: expiresAt,
	}); err != nil {
		return "", time.Time{}, err
	}
	return raw, expiresAt, nil
}

// buildResult issues an access token for the user and assembles the AuthResult.
func (s *service) buildResult(user users.User, rawRefresh string, refreshExp time.Time) (AuthResult, error) {
	now := s.clk.Now()
	access, err := IssueAccessToken(s.cfg.PrivateKey, authctx.Principal{UserID: user.ID}, now, s.cfg.AccessTTL)
	if err != nil {
		return AuthResult{}, err
	}
	return AuthResult{
		User:             user,
		AccessToken:      access,
		AccessExpiresAt:  now.Add(s.cfg.AccessTTL),
		RefreshToken:     rawRefresh,
		RefreshExpiresAt: refreshExp,
	}, nil
}

// errInvalidCredentials is the single response for any failed login so the
// server never reveals whether the email or the password was wrong.
func errInvalidCredentials() error {
	return errors.Unauthenticated("INVALID_CREDENTIALS", "invalid email or password")
}

// validateEmail enforces a minimal, well-formed email check.
func validateEmail(email string) error {
	if email == "" || !strings.Contains(email, "@") {
		return errors.Validation("INVALID_EMAIL", "a valid email is required")
	}
	return nil
}

// validateName enforces the display-name length rules.
func validateName(name string) error {
	if name == "" || len(name) > 200 {
		return errors.Validation("INVALID_NAME", "name must be between 1 and 200 characters")
	}
	return nil
}

// validatePassword enforces the minimum password strength.
func validatePassword(password string) error {
	if len(password) < passwordMinLen {
		return errors.Validation("WEAK_PASSWORD", "password must be at least 8 characters")
	}
	return nil
}
