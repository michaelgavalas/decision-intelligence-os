package users_test

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/db"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/users"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/authctx"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/clock"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/errors"
)

// fakeRepo is an in-memory Repository used to test the service without a
// database. Each field lets a test observe or override behavior.
type fakeRepo struct {
	created    users.CreateParams
	createErr  error
	updateName string
	updateErr  error
	getErr     error
	user       users.User
}

func (f *fakeRepo) Create(_ context.Context, _ db.Querier, p users.CreateParams) (users.User, error) {
	f.created = p
	if f.createErr != nil {
		return users.User{}, f.createErr
	}
	return users.User{ID: p.ID, Email: p.Email, Name: p.Name, PasswordHash: p.PasswordHash}, nil
}

func (f *fakeRepo) GetByID(_ context.Context, _ db.Querier, id uuid.UUID) (users.User, error) {
	if f.getErr != nil {
		return users.User{}, f.getErr
	}
	return users.User{ID: id, Name: f.user.Name}, nil
}

func (f *fakeRepo) GetByEmail(_ context.Context, _ db.Querier, email string) (users.User, error) {
	if f.getErr != nil {
		return users.User{}, f.getErr
	}
	return users.User{Email: email}, nil
}

func (f *fakeRepo) UpdateName(_ context.Context, _ db.Querier, id uuid.UUID, name string) (users.User, error) {
	f.updateName = name
	if f.updateErr != nil {
		return users.User{}, f.updateErr
	}
	return users.User{ID: id, Name: name}, nil
}

// newService builds a service with a fake repository. The pool and tx manager
// are nil because the fake never touches them.
func newService(repo users.Repository) users.Service {
	return users.NewService(nil, nil, repo, clock.Fixed{})
}

func ctxWith(userID uuid.UUID) context.Context {
	return authctx.WithPrincipal(context.Background(), authctx.Principal{UserID: userID})
}

func TestProvision(t *testing.T) {
	tests := []struct {
		name     string
		params   users.ProvisionParams
		wantCode string
	}{
		{
			name:   "valid input creates user",
			params: users.ProvisionParams{Email: "a@example.com", Name: "Ada", PasswordHash: "hash"},
		},
		{
			name:     "empty email is rejected",
			params:   users.ProvisionParams{Email: "  ", Name: "Ada", PasswordHash: "hash"},
			wantCode: "EMAIL_REQUIRED",
		},
		{
			name:     "empty name is rejected",
			params:   users.ProvisionParams{Email: "a@example.com", Name: "", PasswordHash: "hash"},
			wantCode: "NAME_REQUIRED",
		},
		{
			name:     "over-long name is rejected",
			params:   users.ProvisionParams{Email: "a@example.com", Name: strings.Repeat("x", 201), PasswordHash: "hash"},
			wantCode: "NAME_TOO_LONG",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeRepo{}
			svc := newService(repo)

			u, err := svc.Provision(context.Background(), nil, tc.params)

			if tc.wantCode != "" {
				assertCode(t, err, errors.KindValidation, tc.wantCode)
				return
			}
			if err != nil {
				t.Fatalf("Provision: unexpected error: %v", err)
			}
			if u.ID == uuid.Nil {
				t.Error("Provision: expected a generated id")
			}
			if repo.created.Email != "a@example.com" {
				t.Errorf("Provision: repo got email %q, want trimmed", repo.created.Email)
			}
		})
	}
}

func TestGetByIDRequiresAuth(t *testing.T) {
	svc := newService(&fakeRepo{})

	_, err := svc.GetByID(context.Background(), uuid.New())

	assertCode(t, err, errors.KindUnauthenticated, "UNAUTHENTICATED")
}

func TestGetByIDAuthenticated(t *testing.T) {
	caller := uuid.New()
	svc := newService(&fakeRepo{user: users.User{Name: "Ada"}})

	u, err := svc.GetByID(ctxWith(caller), caller)
	if err != nil {
		t.Fatalf("GetByID: unexpected error: %v", err)
	}
	if u.ID != caller {
		t.Errorf("GetByID: id = %v, want %v", u.ID, caller)
	}
}

func TestUpdateProfile(t *testing.T) {
	owner := uuid.New()
	other := uuid.New()

	tests := []struct {
		name     string
		ctx      context.Context
		target   uuid.UUID
		newName  string
		wantKind errors.Kind
		wantCode string
	}{
		{
			name:    "owner updates own name",
			ctx:     ctxWith(owner),
			target:  owner,
			newName: "Grace",
		},
		{
			name:     "no principal is unauthenticated",
			ctx:      context.Background(),
			target:   owner,
			newName:  "Grace",
			wantKind: errors.KindUnauthenticated,
			wantCode: "UNAUTHENTICATED",
		},
		{
			name:     "non-owner is forbidden",
			ctx:      ctxWith(other),
			target:   owner,
			newName:  "Grace",
			wantKind: errors.KindForbidden,
			wantCode: "FORBIDDEN",
		},
		{
			name:     "empty name is rejected",
			ctx:      ctxWith(owner),
			target:   owner,
			newName:  "   ",
			wantKind: errors.KindValidation,
			wantCode: "NAME_REQUIRED",
		},
		{
			name:     "over-long name is rejected",
			ctx:      ctxWith(owner),
			target:   owner,
			newName:  strings.Repeat("y", 201),
			wantKind: errors.KindValidation,
			wantCode: "NAME_TOO_LONG",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeRepo{}
			svc := newService(repo)

			u, err := svc.UpdateProfile(tc.ctx, tc.target, tc.newName)

			if tc.wantCode != "" {
				assertCode(t, err, tc.wantKind, tc.wantCode)
				return
			}
			if err != nil {
				t.Fatalf("UpdateProfile: unexpected error: %v", err)
			}
			if u.Name != "Grace" {
				t.Errorf("UpdateProfile: name = %q, want Grace", u.Name)
			}
			if repo.updateName != "Grace" {
				t.Errorf("UpdateProfile: repo got name %q, want trimmed Grace", repo.updateName)
			}
		})
	}
}

// assertCode fails the test unless err is an *errors.Error with the expected
// kind and code.
func assertCode(t *testing.T, err error, kind errors.Kind, code string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error with code %q, got nil", code)
	}
	if got := errors.KindOf(err); got != kind {
		t.Errorf("error kind = %v, want %v (err: %v)", got, kind, err)
	}
	if got := errors.CodeOf(err); got != code {
		t.Errorf("error code = %q, want %q (err: %v)", got, code, err)
	}
}

func (f *fakeRepo) ListByIDs(_ context.Context, _ db.Querier, _ []uuid.UUID) ([]users.User, error) {
	return nil, nil
}
