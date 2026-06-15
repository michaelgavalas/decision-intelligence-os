package predictions

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

// fakeDecisions is an in-memory Decisions dependency.
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
	byID map[uuid.UUID]Prediction
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{byID: map[uuid.UUID]Prediction{}}
}

func (f *fakeRepo) Create(_ context.Context, _ db.Querier, p Prediction) (Prediction, error) {
	p.CreatedAt = time.Now()
	p.UpdatedAt = p.CreatedAt
	f.byID[p.ID] = p
	return p, nil
}

func (f *fakeRepo) GetByID(_ context.Context, _ db.Querier, id uuid.UUID) (Prediction, error) {
	p, ok := f.byID[id]
	if !ok {
		return Prediction{}, errors.NotFound("PREDICTION_NOT_FOUND", "prediction not found")
	}
	return p, nil
}

func (f *fakeRepo) ListByDecision(_ context.Context, _ db.Querier, _ uuid.UUID) ([]Prediction, error) {
	return nil, nil
}

func (f *fakeRepo) ListByDecisionIDs(_ context.Context, _ db.Querier, _ []uuid.UUID) ([]Prediction, error) {
	return nil, nil
}

func (f *fakeRepo) Update(_ context.Context, _ db.Querier, id uuid.UUID, statement string, probability float64, resolvesAt *time.Time) (Prediction, error) {
	p, ok := f.byID[id]
	if !ok {
		return Prediction{}, errors.NotFound("PREDICTION_NOT_FOUND", "prediction not found")
	}
	p.Statement = statement
	p.Probability = probability
	p.ResolvesAt = resolvesAt
	f.byID[id] = p
	return p, nil
}

func newTestService(repo Repository, rec recorder, dec Decisions) *service {
	return &service{
		pool:      nil,
		tx:        fakeTx{},
		repo:      repo,
		recorder:  rec,
		decisions: dec,
	}
}

func ctxWithUser(userID uuid.UUID) context.Context {
	return authctx.WithPrincipal(context.Background(), authctx.Principal{UserID: userID})
}

func TestCreateRejectsInvalidProbability(t *testing.T) {
	member := uuid.New()
	dec := &fakeDecisions{role: teams.RoleMember}
	s := newTestService(newFakeRepo(), &fakeRecorder{}, dec)

	_, err := s.Create(ctxWithUser(member), CreateInput{DecisionID: uuid.New(), Statement: "It ships", Probability: 1.2})
	if errors.KindOf(err) != errors.KindValidation || errors.CodeOf(err) != "INVALID_PROBABILITY" {
		t.Errorf("err = %v, want Validation/INVALID_PROBABILITY", err)
	}
}

func TestCreateForbiddenForViewer(t *testing.T) {
	viewer := uuid.New()
	dec := &fakeDecisions{role: teams.RoleViewer}
	s := newTestService(newFakeRepo(), &fakeRecorder{}, dec)

	_, err := s.Create(ctxWithUser(viewer), CreateInput{DecisionID: uuid.New(), Statement: "It ships", Probability: 0.6})
	if errors.KindOf(err) != errors.KindForbidden || errors.CodeOf(err) != "PREDICTION_FORBIDDEN" {
		t.Errorf("err = %v, want Forbidden/PREDICTION_FORBIDDEN", err)
	}
}

func TestCreateSucceedsAndRecordsEvent(t *testing.T) {
	member := uuid.New()
	rec := &fakeRecorder{}
	dec := &fakeDecisions{role: teams.RoleMember}
	s := newTestService(newFakeRepo(), rec, dec)

	resolvesAt := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	p, err := s.Create(ctxWithUser(member), CreateInput{DecisionID: uuid.New(), Statement: "It ships on time", Probability: 0.7, ResolvesAt: &resolvesAt})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if p.Probability != 0.7 {
		t.Errorf("probability = %v, want 0.7", p.Probability)
	}
	if p.ResolvesAt == nil || !p.ResolvesAt.Equal(resolvesAt) {
		t.Errorf("ResolvesAt = %v, want %v", p.ResolvesAt, resolvesAt)
	}
	if len(rec.events) != 1 || rec.events[0].Type != events.TypePredictionCreated {
		t.Fatalf("events = %+v, want one PredictionCreated", rec.events)
	}
	if rec.events[0].AggregateType != aggregateType {
		t.Errorf("aggregate type = %q, want %q", rec.events[0].AggregateType, aggregateType)
	}
	if rec.events[0].ActorID == nil || *rec.events[0].ActorID != member {
		t.Errorf("event actor = %v, want %v", rec.events[0].ActorID, member)
	}
}

func TestUpdateForbiddenForViewer(t *testing.T) {
	owner := uuid.New()
	viewer := uuid.New()
	repo := newFakeRepo()
	p, _ := repo.Create(context.Background(), nil, Prediction{ID: uuid.New(), DecisionID: uuid.New(), Statement: "s", Probability: 0.5})
	dec := &fakeDecisions{role: teams.RoleViewer}
	s := newTestService(repo, &fakeRecorder{}, dec)
	_ = owner

	_, err := s.Update(ctxWithUser(viewer), p.ID, "new", 0.4, nil)
	if errors.KindOf(err) != errors.KindForbidden || errors.CodeOf(err) != "PREDICTION_FORBIDDEN" {
		t.Errorf("err = %v, want Forbidden/PREDICTION_FORBIDDEN", err)
	}
}

func (f *fakeRepo) ListByIDs(_ context.Context, _ db.Querier, _ []uuid.UUID) ([]Prediction, error) {
	return nil, nil
}
