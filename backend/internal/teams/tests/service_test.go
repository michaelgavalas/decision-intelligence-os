package teams_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/db"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/teams"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/authctx"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/clock"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/errors"
)

// fakeRepo is an in-memory Repository for testing the service without a
// database. Memberships are keyed by team+user so authorization lookups resolve
// against the seeded state.
type fakeRepo struct {
	memberships map[string]teams.Membership
	adminCount  map[uuid.UUID]int

	createdTeam  teams.Team
	addedRole    string
	addedUser    uuid.UUID
	addMemberErr error
	updatedRole  string
	removed      bool
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		memberships: map[string]teams.Membership{},
		adminCount:  map[uuid.UUID]int{},
	}
}

func key(teamID, userID uuid.UUID) string { return teamID.String() + ":" + userID.String() }

func (f *fakeRepo) seedMember(teamID, userID uuid.UUID, role string) {
	f.memberships[key(teamID, userID)] = teams.Membership{TeamID: teamID, UserID: userID, Role: role}
	if role == teams.RoleAdmin {
		f.adminCount[teamID]++
	}
}

func (f *fakeRepo) CreateTeam(_ context.Context, _ db.Querier, id uuid.UUID, name string) (teams.Team, error) {
	f.createdTeam = teams.Team{ID: id, Name: name}
	return f.createdTeam, nil
}

func (f *fakeRepo) GetTeam(_ context.Context, _ db.Querier, id uuid.UUID) (teams.Team, error) {
	return teams.Team{ID: id}, nil
}

func (f *fakeRepo) AddMember(_ context.Context, _ db.Querier, teamID, userID uuid.UUID, role string) (teams.Membership, error) {
	f.addedUser = userID
	f.addedRole = role
	if f.addMemberErr != nil {
		return teams.Membership{}, f.addMemberErr
	}
	m := teams.Membership{TeamID: teamID, UserID: userID, Role: role}
	f.memberships[key(teamID, userID)] = m
	return m, nil
}

func (f *fakeRepo) GetMembership(_ context.Context, _ db.Querier, teamID, userID uuid.UUID) (teams.Membership, error) {
	m, ok := f.memberships[key(teamID, userID)]
	if !ok {
		return teams.Membership{}, errors.NotFound("MEMBERSHIP_NOT_FOUND", "membership not found")
	}
	return m, nil
}

func (f *fakeRepo) ListMembers(_ context.Context, _ db.Querier, teamID uuid.UUID) ([]teams.Membership, error) {
	var out []teams.Membership
	for _, m := range f.memberships {
		if m.TeamID == teamID {
			out = append(out, m)
		}
	}
	return out, nil
}

func (f *fakeRepo) ListTeamsForUser(_ context.Context, _ db.Querier, userID uuid.UUID) ([]teams.Team, error) {
	var out []teams.Team
	for _, m := range f.memberships {
		if m.UserID == userID {
			out = append(out, teams.Team{ID: m.TeamID})
		}
	}
	return out, nil
}

func (f *fakeRepo) UpdateMemberRole(_ context.Context, _ db.Querier, teamID, userID uuid.UUID, role string) (teams.Membership, error) {
	f.updatedRole = role
	m := teams.Membership{TeamID: teamID, UserID: userID, Role: role}
	f.memberships[key(teamID, userID)] = m
	return m, nil
}

func (f *fakeRepo) RemoveMember(_ context.Context, _ db.Querier, teamID, userID uuid.UUID) error {
	f.removed = true
	delete(f.memberships, key(teamID, userID))
	return nil
}

func (f *fakeRepo) CountAdmins(_ context.Context, _ db.Querier, teamID uuid.UUID) (int, error) {
	return f.adminCount[teamID], nil
}

func newService(repo teams.Repository) teams.Service {
	return teams.NewService(nil, nil, repo, clock.Fixed{})
}

func ctxWith(userID uuid.UUID) context.Context {
	return authctx.WithPrincipal(context.Background(), authctx.Principal{UserID: userID})
}

func TestAddMember(t *testing.T) {
	admin := uuid.New()
	member := uuid.New()
	teamID := uuid.New()
	target := uuid.New()

	t.Run("admin adds member", func(t *testing.T) {
		repo := newFakeRepo()
		repo.seedMember(teamID, admin, teams.RoleAdmin)
		svc := newService(repo)

		m, err := svc.AddMember(ctxWith(admin), teamID, target, teams.RoleMember)
		if err != nil {
			t.Fatalf("AddMember: unexpected error: %v", err)
		}
		if m.Role != teams.RoleMember {
			t.Errorf("AddMember role = %q, want member", m.Role)
		}
	})

	t.Run("non-admin is forbidden", func(t *testing.T) {
		repo := newFakeRepo()
		repo.seedMember(teamID, member, teams.RoleMember)
		svc := newService(repo)

		_, err := svc.AddMember(ctxWith(member), teamID, target, teams.RoleMember)
		assertCode(t, err, errors.KindForbidden, "NOT_TEAM_ADMIN")
	})

	t.Run("non-member is forbidden", func(t *testing.T) {
		repo := newFakeRepo()
		repo.seedMember(teamID, admin, teams.RoleAdmin)
		svc := newService(repo)

		_, err := svc.AddMember(ctxWith(uuid.New()), teamID, target, teams.RoleMember)
		assertCode(t, err, errors.KindForbidden, "NOT_TEAM_MEMBER")
	})

	t.Run("invalid role is rejected", func(t *testing.T) {
		repo := newFakeRepo()
		repo.seedMember(teamID, admin, teams.RoleAdmin)
		svc := newService(repo)

		_, err := svc.AddMember(ctxWith(admin), teamID, target, "superuser")
		assertCode(t, err, errors.KindValidation, "INVALID_ROLE")
	})
}

func TestChangeRoleLastAdmin(t *testing.T) {
	admin := uuid.New()
	teamID := uuid.New()

	t.Run("demoting the last admin is a conflict", func(t *testing.T) {
		repo := newFakeRepo()
		repo.seedMember(teamID, admin, teams.RoleAdmin) // single admin
		svc := newService(repo)

		_, err := svc.ChangeRole(ctxWith(admin), teamID, admin, teams.RoleMember)
		assertCode(t, err, errors.KindConflict, "LAST_ADMIN")
	})

	t.Run("demoting an admin when others remain succeeds", func(t *testing.T) {
		repo := newFakeRepo()
		other := uuid.New()
		repo.seedMember(teamID, admin, teams.RoleAdmin)
		repo.seedMember(teamID, other, teams.RoleAdmin)
		svc := newService(repo)

		_, err := svc.ChangeRole(ctxWith(admin), teamID, other, teams.RoleMember)
		if err != nil {
			t.Fatalf("ChangeRole: unexpected error: %v", err)
		}
		if repo.updatedRole != teams.RoleMember {
			t.Errorf("ChangeRole updated role = %q, want member", repo.updatedRole)
		}
	})
}

func TestRemoveMemberLastAdmin(t *testing.T) {
	admin := uuid.New()
	teamID := uuid.New()

	t.Run("removing the last admin is a conflict", func(t *testing.T) {
		repo := newFakeRepo()
		repo.seedMember(teamID, admin, teams.RoleAdmin)
		svc := newService(repo)

		err := svc.RemoveMember(ctxWith(admin), teamID, admin)
		assertCode(t, err, errors.KindConflict, "LAST_ADMIN")
	})

	t.Run("removing a non-admin member succeeds", func(t *testing.T) {
		repo := newFakeRepo()
		member := uuid.New()
		repo.seedMember(teamID, admin, teams.RoleAdmin)
		repo.seedMember(teamID, member, teams.RoleMember)
		svc := newService(repo)

		if err := svc.RemoveMember(ctxWith(admin), teamID, member); err != nil {
			t.Fatalf("RemoveMember: unexpected error: %v", err)
		}
		if !repo.removed {
			t.Error("RemoveMember: expected repo.RemoveMember to be called")
		}
	})
}

func TestListMembersRequiresMembership(t *testing.T) {
	teamID := uuid.New()
	repo := newFakeRepo()
	repo.seedMember(teamID, uuid.New(), teams.RoleAdmin)
	svc := newService(repo)

	_, err := svc.ListMembers(ctxWith(uuid.New()), teamID)
	assertCode(t, err, errors.KindForbidden, "NOT_TEAM_MEMBER")
}

func TestProvisionPersonalTeamMakesOwnerAdmin(t *testing.T) {
	owner := uuid.New()
	repo := newFakeRepo()
	svc := newService(repo)

	team, err := svc.ProvisionPersonalTeam(context.Background(), nil, owner, "Acme")
	if err != nil {
		t.Fatalf("ProvisionPersonalTeam: unexpected error: %v", err)
	}
	if team.Name != "Acme" {
		t.Errorf("team name = %q, want Acme", team.Name)
	}
	if repo.addedUser != owner || repo.addedRole != teams.RoleAdmin {
		t.Errorf("AddMember got (%v, %q), want owner as admin", repo.addedUser, repo.addedRole)
	}
}

func TestProvisionPersonalTeamValidatesName(t *testing.T) {
	repo := newFakeRepo()
	svc := newService(repo)

	_, err := svc.ProvisionPersonalTeam(context.Background(), nil, uuid.New(), "   ")
	assertCode(t, err, errors.KindValidation, "TEAM_NAME_REQUIRED")
}

func TestListMyTeamsRequiresAuth(t *testing.T) {
	svc := newService(newFakeRepo())

	_, err := svc.ListMyTeams(context.Background())
	assertCode(t, err, errors.KindUnauthenticated, "UNAUTHENTICATED")
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

func (f *fakeRepo) ListByIDs(_ context.Context, _ db.Querier, _ []uuid.UUID) ([]teams.Team, error) {
	return nil, nil
}

func (f *fakeRepo) ListMembersByTeamIDs(_ context.Context, _ db.Querier, _ []uuid.UUID) ([]teams.Membership, error) {
	return nil, nil
}
