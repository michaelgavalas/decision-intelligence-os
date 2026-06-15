package auth

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/db"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/teams"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/users"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/clock"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/errors"
)

// fakeTx runs the work function with a nil querier, so the service's
// transactional flows run without a database.
type fakeTx struct{}

func (fakeTx) WithinTx(ctx context.Context, fn func(q db.Querier) error) error {
	return fn(nil)
}

// fakeUsers is an in-memory UserService.
type fakeUsers struct {
	byEmail map[string]users.User
	byID    map[uuid.UUID]users.User

	provisionErr  error
	provisionedAt int
}

func newFakeUsers() *fakeUsers {
	return &fakeUsers{byEmail: map[string]users.User{}, byID: map[uuid.UUID]users.User{}}
}

func (f *fakeUsers) Provision(_ context.Context, _ db.Querier, p users.ProvisionParams) (users.User, error) {
	f.provisionedAt++
	if f.provisionErr != nil {
		return users.User{}, f.provisionErr
	}
	u := users.User{ID: uuid.New(), Email: p.Email, Name: p.Name, PasswordHash: p.PasswordHash}
	f.byEmail[p.Email] = u
	f.byID[u.ID] = u
	return u, nil
}

func (f *fakeUsers) FindByEmail(_ context.Context, email string) (users.User, error) {
	u, ok := f.byEmail[email]
	if !ok {
		return users.User{}, errors.NotFound("USER_NOT_FOUND", "user not found")
	}
	return u, nil
}

func (f *fakeUsers) GetByID(_ context.Context, id uuid.UUID) (users.User, error) {
	u, ok := f.byID[id]
	if !ok {
		return users.User{}, errors.NotFound("USER_NOT_FOUND", "user not found")
	}
	return u, nil
}

// fakeTeams is an in-memory TeamService.
type fakeTeams struct {
	provisionedOwner uuid.UUID
	provisionedName  string
	calls            int
}

func (f *fakeTeams) ProvisionPersonalTeam(_ context.Context, _ db.Querier, ownerID uuid.UUID, name string) (teams.Team, error) {
	f.calls++
	f.provisionedOwner = ownerID
	f.provisionedName = name
	return teams.Team{ID: uuid.New(), Name: name}, nil
}

// fakeRefresh is an in-memory Repository.
type fakeRefresh struct {
	byHash map[string]RefreshToken
	byID   map[uuid.UUID]RefreshToken

	stored            []RefreshToken
	revokedAllForUser *uuid.UUID
	markReplacedFrom  *uuid.UUID
	markReplacedTo    *uuid.UUID
	revokedID         *uuid.UUID
}

func newFakeRefresh() *fakeRefresh {
	return &fakeRefresh{byHash: map[string]RefreshToken{}, byID: map[uuid.UUID]RefreshToken{}}
}

func (f *fakeRefresh) Store(_ context.Context, _ db.Querier, t RefreshToken) (RefreshToken, error) {
	f.stored = append(f.stored, t)
	f.byHash[t.TokenHash] = t
	f.byID[t.ID] = t
	return t, nil
}

func (f *fakeRefresh) GetByHash(_ context.Context, _ db.Querier, hash string) (RefreshToken, error) {
	t, ok := f.byHash[hash]
	if !ok {
		return RefreshToken{}, errors.NotFound("REFRESH_NOT_FOUND", "refresh token not found")
	}
	return t, nil
}

func (f *fakeRefresh) Revoke(_ context.Context, _ db.Querier, id uuid.UUID) error {
	f.revokedID = &id
	if t, ok := f.byID[id]; ok {
		now := time.Now()
		t.RevokedAt = &now
		f.byID[id] = t
		f.byHash[t.TokenHash] = t
	}
	return nil
}

func (f *fakeRefresh) RevokeAllForUser(_ context.Context, _ db.Querier, userID uuid.UUID) error {
	f.revokedAllForUser = &userID
	return nil
}

func (f *fakeRefresh) MarkReplaced(_ context.Context, _ db.Querier, id, replacedBy uuid.UUID) error {
	f.markReplacedFrom = &id
	f.markReplacedTo = &replacedBy
	return nil
}

// newTestService builds a service backed by the supplied fakes and a real key
// pair, so issued access tokens are valid.
func newTestService(t *testing.T, u UserService, tm TeamService, r Repository, lim Limiter) *service {
	t.Helper()
	pub, priv, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}
	return &service{
		pool:    nil,
		tx:      fakeTx{},
		users:   u,
		teams:   tm,
		refresh: r,
		limiter: lim,
		cfg: TokenConfig{
			PrivateKey: priv,
			PublicKey:  pub,
			AccessTTL:  time.Hour,
			RefreshTTL: 24 * time.Hour,
		},
		clk: clock.Fixed{T: time.Date(2026, 6, 14, 12, 0, 0, 0, time.UTC)},
	}
}

// allowLimiter always permits.
type allowLimiter struct{ called bool }

func (l *allowLimiter) Allow(context.Context, string, int, time.Duration) (bool, error) {
	l.called = true
	return true, nil
}

// denyLimiter always denies.
type denyLimiter struct{}

func (denyLimiter) Allow(context.Context, string, int, time.Duration) (bool, error) {
	return false, nil
}

func TestRegisterSuccess(t *testing.T) {
	fu, ft, fr := newFakeUsers(), &fakeTeams{}, newFakeRefresh()
	s := newTestService(t, fu, ft, fr, &allowLimiter{})

	res, err := s.Register(context.Background(), RegisterInput{
		Email: "a@example.com", Name: "Ada", Password: "supersecret",
	})
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if res.AccessToken == "" || res.RefreshToken == "" {
		t.Errorf("tokens empty: access=%q refresh=%q", res.AccessToken, res.RefreshToken)
	}
	if fu.provisionedAt != 1 {
		t.Errorf("Provision calls = %d, want 1", fu.provisionedAt)
	}
	if ft.calls != 1 || ft.provisionedName != "Ada's Team" {
		t.Errorf("team provision = (%d, %q), want (1, \"Ada's Team\")", ft.calls, ft.provisionedName)
	}
	if len(fr.stored) != 1 {
		t.Errorf("stored refresh tokens = %d, want 1", len(fr.stored))
	}
	// Access token round-trips to the new user.
	p, err := s.ParseAccessToken(res.AccessToken)
	if err != nil || p.UserID != res.User.ID {
		t.Errorf("ParseAccessToken = (%v, %v), want user %v", p, err, res.User.ID)
	}
}

func TestRegisterValidation(t *testing.T) {
	cases := map[string]struct {
		in   RegisterInput
		code string
	}{
		"bad email":     {RegisterInput{Email: "noat", Name: "Ada", Password: "supersecret"}, "INVALID_EMAIL"},
		"empty name":    {RegisterInput{Email: "a@b.com", Name: "", Password: "supersecret"}, "INVALID_NAME"},
		"weak password": {RegisterInput{Email: "a@b.com", Name: "Ada", Password: "short"}, "WEAK_PASSWORD"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			s := newTestService(t, newFakeUsers(), &fakeTeams{}, newFakeRefresh(), &allowLimiter{})
			_, err := s.Register(context.Background(), tc.in)
			if errors.KindOf(err) != errors.KindValidation || errors.CodeOf(err) != tc.code {
				t.Errorf("err = %v, want Validation/%s", err, tc.code)
			}
		})
	}
}

func TestRegisterEmailTakenPropagates(t *testing.T) {
	fu := newFakeUsers()
	fu.provisionErr = errors.Conflict("EMAIL_TAKEN", "email already registered")
	s := newTestService(t, fu, &fakeTeams{}, newFakeRefresh(), &allowLimiter{})

	_, err := s.Register(context.Background(), RegisterInput{
		Email: "a@example.com", Name: "Ada", Password: "supersecret",
	})
	if errors.KindOf(err) != errors.KindConflict || errors.CodeOf(err) != "EMAIL_TAKEN" {
		t.Errorf("err = %v, want Conflict/EMAIL_TAKEN", err)
	}
}

func TestLoginWrongPassword(t *testing.T) {
	fu := newFakeUsers()
	hash, _ := HashPassword("correct password")
	u := users.User{ID: uuid.New(), Email: "a@example.com", PasswordHash: hash}
	fu.byEmail[u.Email] = u
	fu.byID[u.ID] = u
	s := newTestService(t, fu, &fakeTeams{}, newFakeRefresh(), &allowLimiter{})

	_, err := s.Login(context.Background(), "a@example.com", "wrong password", "ip")
	if errors.KindOf(err) != errors.KindUnauthenticated || errors.CodeOf(err) != "INVALID_CREDENTIALS" {
		t.Errorf("err = %v, want Unauthenticated/INVALID_CREDENTIALS", err)
	}
}

func TestLoginUnknownEmail(t *testing.T) {
	s := newTestService(t, newFakeUsers(), &fakeTeams{}, newFakeRefresh(), &allowLimiter{})

	_, err := s.Login(context.Background(), "missing@example.com", "whatever123", "ip")
	if errors.KindOf(err) != errors.KindUnauthenticated || errors.CodeOf(err) != "INVALID_CREDENTIALS" {
		t.Errorf("err = %v, want Unauthenticated/INVALID_CREDENTIALS (no leak)", err)
	}
}

func TestLoginSuccess(t *testing.T) {
	fu := newFakeUsers()
	hash, _ := HashPassword("correct password")
	u := users.User{ID: uuid.New(), Email: "a@example.com", PasswordHash: hash}
	fu.byEmail[u.Email] = u
	fu.byID[u.ID] = u
	fr := newFakeRefresh()
	s := newTestService(t, fu, &fakeTeams{}, fr, &allowLimiter{})

	res, err := s.Login(context.Background(), "a@example.com", "correct password", "ip")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if res.AccessToken == "" || res.RefreshToken == "" {
		t.Error("Login returned empty tokens")
	}
	if len(fr.stored) != 1 {
		t.Errorf("stored refresh = %d, want 1", len(fr.stored))
	}
}

func TestLoginRateLimited(t *testing.T) {
	fu := newFakeUsers()
	s := newTestService(t, fu, &fakeTeams{}, newFakeRefresh(), denyLimiter{})

	_, err := s.Login(context.Background(), "a@example.com", "whatever123", "ip")
	if errors.KindOf(err) != errors.KindUnauthenticated || errors.CodeOf(err) != "RATE_LIMITED" {
		t.Errorf("err = %v, want Unauthenticated/RATE_LIMITED", err)
	}
	if fu.provisionedAt != 0 {
		t.Error("FindByEmail-side work ran despite rate limit")
	}
}

func TestRefreshHappyPath(t *testing.T) {
	fu := newFakeUsers()
	userID := uuid.New()
	fu.byID[userID] = users.User{ID: userID, Email: "a@example.com"}
	fr := newFakeRefresh()
	s := newTestService(t, fu, &fakeTeams{}, fr, &allowLimiter{})

	// Seed a valid, unrevoked, unexpired token.
	raw, hash, _ := GenerateRefreshToken()
	old := RefreshToken{
		ID: uuid.New(), UserID: userID, TokenHash: hash,
		ExpiresAt: s.clk.Now().Add(time.Hour),
	}
	fr.byHash[hash], fr.byID[old.ID] = old, old

	res, err := s.Refresh(context.Background(), raw)
	if err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	if res.RefreshToken == "" || res.RefreshToken == raw {
		t.Error("Refresh did not issue a new refresh token")
	}
	if len(fr.stored) != 1 {
		t.Errorf("new tokens stored = %d, want 1", len(fr.stored))
	}
	if fr.markReplacedFrom == nil || *fr.markReplacedFrom != old.ID {
		t.Errorf("MarkReplaced from = %v, want %v", fr.markReplacedFrom, old.ID)
	}
	if fr.markReplacedTo == nil || *fr.markReplacedTo != fr.stored[0].ID {
		t.Errorf("MarkReplaced to = %v, want %v", fr.markReplacedTo, fr.stored[0].ID)
	}
}

func TestRefreshReuseDetected(t *testing.T) {
	userID := uuid.New()
	fr := newFakeRefresh()
	s := newTestService(t, newFakeUsers(), &fakeTeams{}, fr, &allowLimiter{})

	raw, hash, _ := GenerateRefreshToken()
	revokedAt := s.clk.Now()
	old := RefreshToken{
		ID: uuid.New(), UserID: userID, TokenHash: hash,
		ExpiresAt: s.clk.Now().Add(time.Hour), RevokedAt: &revokedAt,
	}
	fr.byHash[hash], fr.byID[old.ID] = old, old

	_, err := s.Refresh(context.Background(), raw)
	if errors.KindOf(err) != errors.KindUnauthenticated || errors.CodeOf(err) != "TOKEN_REUSE" {
		t.Errorf("err = %v, want Unauthenticated/TOKEN_REUSE", err)
	}
	if fr.revokedAllForUser == nil || *fr.revokedAllForUser != userID {
		t.Errorf("RevokeAllForUser = %v, want %v", fr.revokedAllForUser, userID)
	}
}

func TestRefreshExpired(t *testing.T) {
	fr := newFakeRefresh()
	s := newTestService(t, newFakeUsers(), &fakeTeams{}, fr, &allowLimiter{})

	raw, hash, _ := GenerateRefreshToken()
	old := RefreshToken{
		ID: uuid.New(), UserID: uuid.New(), TokenHash: hash,
		ExpiresAt: s.clk.Now().Add(-time.Hour),
	}
	fr.byHash[hash], fr.byID[old.ID] = old, old

	_, err := s.Refresh(context.Background(), raw)
	if errors.KindOf(err) != errors.KindUnauthenticated || errors.CodeOf(err) != "EXPIRED_REFRESH" {
		t.Errorf("err = %v, want Unauthenticated/EXPIRED_REFRESH", err)
	}
}

func TestRefreshUnknownHash(t *testing.T) {
	s := newTestService(t, newFakeUsers(), &fakeTeams{}, newFakeRefresh(), &allowLimiter{})

	_, err := s.Refresh(context.Background(), "nonexistent-raw-token")
	if errors.KindOf(err) != errors.KindUnauthenticated || errors.CodeOf(err) != "INVALID_REFRESH" {
		t.Errorf("err = %v, want Unauthenticated/INVALID_REFRESH", err)
	}
}

func TestLogout(t *testing.T) {
	fr := newFakeRefresh()
	s := newTestService(t, newFakeUsers(), &fakeTeams{}, fr, &allowLimiter{})

	raw, hash, _ := GenerateRefreshToken()
	tok := RefreshToken{ID: uuid.New(), UserID: uuid.New(), TokenHash: hash, ExpiresAt: s.clk.Now().Add(time.Hour)}
	fr.byHash[hash], fr.byID[tok.ID] = tok, tok

	if err := s.Logout(context.Background(), raw); err != nil {
		t.Fatalf("Logout: %v", err)
	}
	if fr.revokedID == nil || *fr.revokedID != tok.ID {
		t.Errorf("Revoke id = %v, want %v", fr.revokedID, tok.ID)
	}
}

func TestLogoutUnknownTokenIsNil(t *testing.T) {
	fr := newFakeRefresh()
	s := newTestService(t, newFakeUsers(), &fakeTeams{}, fr, &allowLimiter{})

	if err := s.Logout(context.Background(), "missing"); err != nil {
		t.Errorf("Logout unknown = %v, want nil", err)
	}
	if fr.revokedID != nil {
		t.Error("Logout revoked something for an unknown token")
	}
}
