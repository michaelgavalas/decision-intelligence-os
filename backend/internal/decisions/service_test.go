package decisions

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/db"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/events"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/teams"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/authctx"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/clock"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/errors"
)

// fakeTx runs the work function with a nil querier, so transactional flows run
// without a database.
type fakeTx struct{}

func (fakeTx) WithinTx(_ context.Context, fn func(q db.Querier) error) error {
	return fn(nil)
}

// fakeRecorder captures recorded events.
type fakeRecorder struct {
	events []events.Event
}

func (f *fakeRecorder) Record(_ context.Context, _ db.Querier, e events.Event) error {
	f.events = append(f.events, e)
	return nil
}

// fakeTeams is an in-memory Teams dependency keyed by (teamID, userID).
type fakeTeams struct {
	roles map[uuid.UUID]string // userID -> role within the configured team
}

func (f *fakeTeams) GetMembership(_ context.Context, teamID, userID uuid.UUID) (teams.Membership, error) {
	role, ok := f.roles[userID]
	if !ok {
		return teams.Membership{}, errors.Forbidden("NOT_TEAM_MEMBER", "caller is not a member of the team")
	}
	return teams.Membership{TeamID: teamID, UserID: userID, Role: role}, nil
}

// fakeRepo is an in-memory Repository.
type fakeRepo struct {
	byID map[uuid.UUID]Decision

	createErr error
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{byID: map[uuid.UUID]Decision{}}
}

func (f *fakeRepo) Create(_ context.Context, _ db.Querier, d Decision) (Decision, error) {
	if f.createErr != nil {
		return Decision{}, f.createErr
	}
	d.CreatedAt = time.Now()
	d.UpdatedAt = d.CreatedAt
	f.byID[d.ID] = d
	return d, nil
}

func (f *fakeRepo) GetByID(_ context.Context, _ db.Querier, id uuid.UUID) (Decision, error) {
	d, ok := f.byID[id]
	if !ok {
		return Decision{}, errors.NotFound("DECISION_NOT_FOUND", "decision not found")
	}
	return d, nil
}

func (f *fakeRepo) ListByTeam(_ context.Context, _ db.Querier, _ uuid.UUID, _ int) ([]Decision, error) {
	return nil, nil
}

func (f *fakeRepo) ListByTeamAfter(_ context.Context, _ db.Querier, _ uuid.UUID, _ time.Time, _ uuid.UUID, _ int) ([]Decision, error) {
	return nil, nil
}

func (f *fakeRepo) CountByTeam(_ context.Context, _ db.Querier, _ uuid.UUID) (int, error) {
	return len(f.byID), nil
}

func (f *fakeRepo) Update(_ context.Context, _ db.Querier, id uuid.UUID, title, description string) (Decision, error) {
	d, ok := f.byID[id]
	if !ok {
		return Decision{}, errors.NotFound("DECISION_NOT_FOUND", "decision not found")
	}
	d.Title = title
	d.Description = description
	f.byID[id] = d
	return d, nil
}

func (f *fakeRepo) UpdateStatus(_ context.Context, _ db.Querier, id uuid.UUID, status string, decidedAt *time.Time) (Decision, error) {
	d, ok := f.byID[id]
	if !ok {
		return Decision{}, errors.NotFound("DECISION_NOT_FOUND", "decision not found")
	}
	d.Status = status
	d.DecidedAt = decidedAt
	f.byID[id] = d
	return d, nil
}

// newTestService builds a service backed by the supplied fakes.
func newTestService(repo Repository, rec recorder, tm Teams) *service {
	return &service{
		pool:     nil,
		tx:       fakeTx{},
		repo:     repo,
		recorder: rec,
		teams:    tm,
		clk:      clock.Fixed{T: time.Date(2026, 6, 14, 12, 0, 0, 0, time.UTC)},
	}
}

func ctxWithUser(userID uuid.UUID) context.Context {
	return authctx.WithPrincipal(context.Background(), authctx.Principal{UserID: userID})
}

func TestCreateForbiddenForViewer(t *testing.T) {
	viewer := uuid.New()
	teamID := uuid.New()
	tm := &fakeTeams{roles: map[uuid.UUID]string{viewer: teams.RoleViewer}}
	s := newTestService(newFakeRepo(), &fakeRecorder{}, tm)

	_, err := s.Create(ctxWithUser(viewer), CreateInput{TeamID: teamID, Title: "X"})
	if errors.KindOf(err) != errors.KindForbidden || errors.CodeOf(err) != "DECISION_CREATE_FORBIDDEN" {
		t.Errorf("err = %v, want Forbidden/DECISION_CREATE_FORBIDDEN", err)
	}
}

func TestCreateForbiddenForNonMember(t *testing.T) {
	stranger := uuid.New()
	tm := &fakeTeams{roles: map[uuid.UUID]string{}}
	s := newTestService(newFakeRepo(), &fakeRecorder{}, tm)

	_, err := s.Create(ctxWithUser(stranger), CreateInput{TeamID: uuid.New(), Title: "X"})
	if errors.KindOf(err) != errors.KindForbidden {
		t.Errorf("err = %v, want Forbidden", err)
	}
}

func TestCreateRequiresTitle(t *testing.T) {
	member := uuid.New()
	tm := &fakeTeams{roles: map[uuid.UUID]string{member: teams.RoleMember}}
	s := newTestService(newFakeRepo(), &fakeRecorder{}, tm)

	_, err := s.Create(ctxWithUser(member), CreateInput{TeamID: uuid.New(), Title: "   "})
	if errors.KindOf(err) != errors.KindValidation || errors.CodeOf(err) != "DECISION_TITLE_REQUIRED" {
		t.Errorf("err = %v, want Validation/DECISION_TITLE_REQUIRED", err)
	}
}

func TestCreateSucceedsAndRecordsEvent(t *testing.T) {
	member := uuid.New()
	teamID := uuid.New()
	tm := &fakeTeams{roles: map[uuid.UUID]string{member: teams.RoleMember}}
	rec := &fakeRecorder{}
	s := newTestService(newFakeRepo(), rec, tm)

	d, err := s.Create(ctxWithUser(member), CreateInput{TeamID: teamID, Title: "Launch in EU", Description: "x"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if d.Status != StatusDraft {
		t.Errorf("status = %q, want draft", d.Status)
	}
	if d.OwnerID != member {
		t.Errorf("owner = %v, want %v", d.OwnerID, member)
	}
	if len(rec.events) != 1 || rec.events[0].Type != events.TypeDecisionCreated {
		t.Fatalf("events = %+v, want one DecisionCreated", rec.events)
	}
	if rec.events[0].ActorID == nil || *rec.events[0].ActorID != member {
		t.Errorf("event actor = %v, want %v", rec.events[0].ActorID, member)
	}
}

func TestUpdateForbiddenForNonOwnerNonAdmin(t *testing.T) {
	owner := uuid.New()
	other := uuid.New()
	teamID := uuid.New()
	repo := newFakeRepo()
	d, _ := repo.Create(context.Background(), nil, Decision{ID: uuid.New(), TeamID: teamID, OwnerID: owner, Title: "T", Status: StatusDraft})
	tm := &fakeTeams{roles: map[uuid.UUID]string{owner: teams.RoleMember, other: teams.RoleMember}}
	s := newTestService(repo, &fakeRecorder{}, tm)

	_, err := s.Update(ctxWithUser(other), d.ID, UpdateInput{Title: "New"})
	if errors.KindOf(err) != errors.KindForbidden || errors.CodeOf(err) != "DECISION_FORBIDDEN" {
		t.Errorf("err = %v, want Forbidden/DECISION_FORBIDDEN", err)
	}
}

func TestUpdateAllowedForTeamAdmin(t *testing.T) {
	owner := uuid.New()
	admin := uuid.New()
	teamID := uuid.New()
	repo := newFakeRepo()
	d, _ := repo.Create(context.Background(), nil, Decision{ID: uuid.New(), TeamID: teamID, OwnerID: owner, Title: "T", Status: StatusDraft})
	tm := &fakeTeams{roles: map[uuid.UUID]string{owner: teams.RoleMember, admin: teams.RoleAdmin}}
	rec := &fakeRecorder{}
	s := newTestService(repo, rec, tm)

	updated, err := s.Update(ctxWithUser(admin), d.ID, UpdateInput{Title: "Revised"})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Title != "Revised" {
		t.Errorf("title = %q, want Revised", updated.Title)
	}
	if len(rec.events) != 1 || rec.events[0].Type != events.TypeDecisionUpdated {
		t.Errorf("events = %+v, want one DecisionUpdated", rec.events)
	}
}

func TestTransitionIllegalIsValidation(t *testing.T) {
	owner := uuid.New()
	teamID := uuid.New()
	repo := newFakeRepo()
	d, _ := repo.Create(context.Background(), nil, Decision{ID: uuid.New(), TeamID: teamID, OwnerID: owner, Title: "T", Status: StatusDraft})
	tm := &fakeTeams{roles: map[uuid.UUID]string{owner: teams.RoleMember}}
	s := newTestService(repo, &fakeRecorder{}, tm)

	// draft -> decided is not allowed.
	_, err := s.Transition(ctxWithUser(owner), d.ID, StatusDecided)
	if errors.KindOf(err) != errors.KindValidation || errors.CodeOf(err) != "INVALID_TRANSITION" {
		t.Errorf("err = %v, want Validation/INVALID_TRANSITION", err)
	}
}

func TestTransitionToDecidedSetsDecidedAt(t *testing.T) {
	owner := uuid.New()
	teamID := uuid.New()
	repo := newFakeRepo()
	d, _ := repo.Create(context.Background(), nil, Decision{ID: uuid.New(), TeamID: teamID, OwnerID: owner, Title: "T", Status: StatusActive})
	tm := &fakeTeams{roles: map[uuid.UUID]string{owner: teams.RoleMember}}
	rec := &fakeRecorder{}
	s := newTestService(repo, rec, tm)

	updated, err := s.Transition(ctxWithUser(owner), d.ID, StatusDecided)
	if err != nil {
		t.Fatalf("Transition: %v", err)
	}
	if updated.Status != StatusDecided {
		t.Errorf("status = %q, want decided", updated.Status)
	}
	if updated.DecidedAt == nil || !updated.DecidedAt.Equal(time.Date(2026, 6, 14, 12, 0, 0, 0, time.UTC)) {
		t.Errorf("DecidedAt = %v, want fixed clock time", updated.DecidedAt)
	}
	if len(rec.events) != 1 || rec.events[0].Type != events.TypeDecisionUpdated {
		t.Errorf("events = %+v, want one DecisionUpdated", rec.events)
	}
}

func TestMarkDecidedNoOpWhenAlreadyDecided(t *testing.T) {
	owner := uuid.New()
	teamID := uuid.New()
	repo := newFakeRepo()
	decidedAt := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	d, _ := repo.Create(context.Background(), nil, Decision{ID: uuid.New(), TeamID: teamID, OwnerID: owner, Title: "T", Status: StatusDecided, DecidedAt: &decidedAt})
	tm := &fakeTeams{roles: map[uuid.UUID]string{owner: teams.RoleMember}}
	rec := &fakeRecorder{}
	s := newTestService(repo, rec, tm)

	got, err := s.MarkDecided(ctxWithUser(owner), nil, d.ID)
	if err != nil {
		t.Fatalf("MarkDecided: %v", err)
	}
	if got.Status != StatusDecided {
		t.Errorf("status = %q, want decided", got.Status)
	}
	if got.DecidedAt == nil || !got.DecidedAt.Equal(decidedAt) {
		t.Errorf("DecidedAt = %v, want %v (unchanged)", got.DecidedAt, decidedAt)
	}
	if len(rec.events) != 0 {
		t.Errorf("events = %+v, want none on no-op", rec.events)
	}
}

func TestMarkDecidedTransitionsActiveAndRecordsEvent(t *testing.T) {
	owner := uuid.New()
	teamID := uuid.New()
	repo := newFakeRepo()
	d, _ := repo.Create(context.Background(), nil, Decision{ID: uuid.New(), TeamID: teamID, OwnerID: owner, Title: "T", Status: StatusActive})
	tm := &fakeTeams{roles: map[uuid.UUID]string{owner: teams.RoleMember}}
	rec := &fakeRecorder{}
	s := newTestService(repo, rec, tm)

	got, err := s.MarkDecided(ctxWithUser(owner), nil, d.ID)
	if err != nil {
		t.Fatalf("MarkDecided: %v", err)
	}
	if got.Status != StatusDecided {
		t.Errorf("status = %q, want decided", got.Status)
	}
	if got.DecidedAt == nil || !got.DecidedAt.Equal(time.Date(2026, 6, 14, 12, 0, 0, 0, time.UTC)) {
		t.Errorf("DecidedAt = %v, want fixed clock time", got.DecidedAt)
	}
	if len(rec.events) != 1 || rec.events[0].Type != events.TypeDecisionUpdated {
		t.Fatalf("events = %+v, want one DecisionUpdated", rec.events)
	}
	if rec.events[0].ActorID == nil || *rec.events[0].ActorID != owner {
		t.Errorf("event actor = %v, want %v", rec.events[0].ActorID, owner)
	}
}

func TestAuthorizeAccessReturnsRole(t *testing.T) {
	member := uuid.New()
	teamID := uuid.New()
	repo := newFakeRepo()
	d, _ := repo.Create(context.Background(), nil, Decision{ID: uuid.New(), TeamID: teamID, OwnerID: member, Title: "T", Status: StatusDraft})
	tm := &fakeTeams{roles: map[uuid.UUID]string{member: teams.RoleMember}}
	s := newTestService(repo, &fakeRecorder{}, tm)

	got, role, err := s.AuthorizeAccess(ctxWithUser(member), d.ID)
	if err != nil {
		t.Fatalf("AuthorizeAccess: %v", err)
	}
	if got.ID != d.ID || role != teams.RoleMember {
		t.Errorf("AuthorizeAccess = (%v, %q), want (%v, member)", got.ID, role, d.ID)
	}
}

func (f *fakeRepo) ListByIDs(_ context.Context, _ db.Querier, _ []uuid.UUID) ([]Decision, error) {
	return nil, nil
}
