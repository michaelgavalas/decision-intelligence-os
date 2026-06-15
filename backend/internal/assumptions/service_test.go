package assumptions

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/decisions"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/db"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/events"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/teams"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/authctx"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/errors"
)

// fakeTx runs the work function with a nil querier.
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

// fakeDecisions is an in-memory Decisions dependency. It returns the configured
// role for a known decision and Forbidden otherwise.
type fakeDecisions struct {
	role      string
	notMember bool
}

func (f *fakeDecisions) AuthorizeAccess(_ context.Context, id uuid.UUID) (decisions.Decision, string, error) {
	if f.notMember {
		return decisions.Decision{}, "", errors.Forbidden("NOT_TEAM_MEMBER", "not a member")
	}
	return decisions.Decision{ID: id}, f.role, nil
}

// fakeRepo is an in-memory Repository.
type fakeRepo struct {
	byID map[uuid.UUID]Assumption
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{byID: map[uuid.UUID]Assumption{}}
}

func (f *fakeRepo) Create(_ context.Context, _ db.Querier, a Assumption) (Assumption, error) {
	a.CreatedAt = time.Now()
	a.UpdatedAt = a.CreatedAt
	f.byID[a.ID] = a
	return a, nil
}

func (f *fakeRepo) GetByID(_ context.Context, _ db.Querier, id uuid.UUID) (Assumption, error) {
	a, ok := f.byID[id]
	if !ok {
		return Assumption{}, errors.NotFound("ASSUMPTION_NOT_FOUND", "assumption not found")
	}
	return a, nil
}

func (f *fakeRepo) ListByDecision(_ context.Context, _ db.Querier, _ uuid.UUID) ([]Assumption, error) {
	return nil, nil
}

func (f *fakeRepo) ListByDecisionIDs(_ context.Context, _ db.Querier, _ []uuid.UUID) ([]Assumption, error) {
	return nil, nil
}

func (f *fakeRepo) Update(_ context.Context, _ db.Querier, id uuid.UUID, statement string, confidence float64) (Assumption, error) {
	a, ok := f.byID[id]
	if !ok {
		return Assumption{}, errors.NotFound("ASSUMPTION_NOT_FOUND", "assumption not found")
	}
	a.Statement = statement
	a.Confidence = confidence
	f.byID[id] = a
	return a, nil
}

func (f *fakeRepo) Delete(_ context.Context, _ db.Querier, id uuid.UUID) error {
	delete(f.byID, id)
	return nil
}

func newTestService(repo Repository, rec recorder, dec Decisions) *service {
	return &service{pool: nil, tx: fakeTx{}, repo: repo, recorder: rec, decisions: dec}
}

func ctxWithUser(userID uuid.UUID) context.Context {
	return authctx.WithPrincipal(context.Background(), authctx.Principal{UserID: userID})
}

func TestAddInvalidConfidence(t *testing.T) {
	user := uuid.New()
	dec := &fakeDecisions{role: teams.RoleMember}
	s := newTestService(newFakeRepo(), &fakeRecorder{}, dec)

	_, err := s.Add(ctxWithUser(user), AddInput{DecisionID: uuid.New(), Statement: "x", Confidence: 1.5})
	if errors.KindOf(err) != errors.KindValidation || errors.CodeOf(err) != "INVALID_CONFIDENCE" {
		t.Errorf("err = %v, want Validation/INVALID_CONFIDENCE", err)
	}
}

func TestAddForbiddenForViewer(t *testing.T) {
	user := uuid.New()
	dec := &fakeDecisions{role: teams.RoleViewer}
	s := newTestService(newFakeRepo(), &fakeRecorder{}, dec)

	_, err := s.Add(ctxWithUser(user), AddInput{DecisionID: uuid.New(), Statement: "x", Confidence: 0.5})
	if errors.KindOf(err) != errors.KindForbidden || errors.CodeOf(err) != "ASSUMPTION_FORBIDDEN" {
		t.Errorf("err = %v, want Forbidden/ASSUMPTION_FORBIDDEN", err)
	}
}

func TestAddForbiddenForNonMember(t *testing.T) {
	user := uuid.New()
	dec := &fakeDecisions{notMember: true}
	s := newTestService(newFakeRepo(), &fakeRecorder{}, dec)

	_, err := s.Add(ctxWithUser(user), AddInput{DecisionID: uuid.New(), Statement: "x", Confidence: 0.5})
	if errors.KindOf(err) != errors.KindForbidden {
		t.Errorf("err = %v, want Forbidden", err)
	}
}

func TestAddSucceedsAndRecordsEvent(t *testing.T) {
	user := uuid.New()
	dec := &fakeDecisions{role: teams.RoleMember}
	rec := &fakeRecorder{}
	s := newTestService(newFakeRepo(), rec, dec)

	a, err := s.Add(ctxWithUser(user), AddInput{DecisionID: uuid.New(), Statement: "Market grows", Confidence: 0.7})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	if a.Confidence != 0.7 {
		t.Errorf("confidence = %v, want 0.7", a.Confidence)
	}
	if len(rec.events) != 1 || rec.events[0].Type != events.TypeAssumptionAdded {
		t.Fatalf("events = %+v, want one AssumptionAdded", rec.events)
	}
	if rec.events[0].ActorID == nil || *rec.events[0].ActorID != user {
		t.Errorf("actor = %v, want %v", rec.events[0].ActorID, user)
	}
}

func TestAuthorizeAccessReturnsRole(t *testing.T) {
	user := uuid.New()
	repo := newFakeRepo()
	a, _ := repo.Create(context.Background(), nil, Assumption{ID: uuid.New(), DecisionID: uuid.New(), Statement: "s", Confidence: 0.4})
	dec := &fakeDecisions{role: teams.RoleAdmin}
	s := newTestService(repo, &fakeRecorder{}, dec)

	got, role, err := s.AuthorizeAccess(ctxWithUser(user), a.ID)
	if err != nil {
		t.Fatalf("AuthorizeAccess: %v", err)
	}
	if got.ID != a.ID || role != teams.RoleAdmin {
		t.Errorf("AuthorizeAccess = (%v, %q), want (%v, admin)", got.ID, role, a.ID)
	}
}

func TestUpdateForbiddenForViewer(t *testing.T) {
	user := uuid.New()
	repo := newFakeRepo()
	a, _ := repo.Create(context.Background(), nil, Assumption{ID: uuid.New(), DecisionID: uuid.New(), Statement: "s", Confidence: 0.4})
	dec := &fakeDecisions{role: teams.RoleViewer}
	s := newTestService(repo, &fakeRecorder{}, dec)

	_, err := s.Update(ctxWithUser(user), a.ID, "new", 0.5)
	if errors.KindOf(err) != errors.KindForbidden {
		t.Errorf("err = %v, want Forbidden", err)
	}
}

func (f *fakeRepo) ListByIDs(_ context.Context, _ db.Querier, _ []uuid.UUID) ([]Assumption, error) {
	return nil, nil
}
