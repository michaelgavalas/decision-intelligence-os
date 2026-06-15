package outcomes

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
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/clock"
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

// fakeDecisions is an in-memory Decisions dependency that records whether
// MarkDecided was invoked.
type fakeDecisions struct {
	decision     decisions.Decision
	role         string
	notMember    bool
	markedCalled bool
}

func (f *fakeDecisions) AuthorizeAccess(_ context.Context, _ uuid.UUID) (decisions.Decision, string, error) {
	if f.notMember {
		return decisions.Decision{}, "", errors.Forbidden("NOT_TEAM_MEMBER", "not a member")
	}
	return f.decision, f.role, nil
}

func (f *fakeDecisions) MarkDecided(_ context.Context, _ db.Querier, _ uuid.UUID) (decisions.Decision, error) {
	f.markedCalled = true
	d := f.decision
	d.Status = decisions.StatusDecided
	return d, nil
}

// fakeRepo is an in-memory Repository whose Upsert is idempotent on decision id.
type fakeRepo struct {
	byDecision map[uuid.UUID]Outcome
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{byDecision: map[uuid.UUID]Outcome{}}
}

func (f *fakeRepo) Upsert(_ context.Context, _ db.Querier, o Outcome) (Outcome, error) {
	if existing, ok := f.byDecision[o.DecisionID]; ok {
		// Mimic ON CONFLICT (decision_id) DO UPDATE: keep the original id/created.
		o.ID = existing.ID
		o.CreatedAt = existing.CreatedAt
	} else {
		o.CreatedAt = time.Now()
	}
	o.UpdatedAt = time.Now()
	f.byDecision[o.DecisionID] = o
	return o, nil
}

func (f *fakeRepo) GetByDecision(_ context.Context, _ db.Querier, decisionID uuid.UUID) (Outcome, error) {
	o, ok := f.byDecision[decisionID]
	if !ok {
		return Outcome{}, errors.NotFound("OUTCOME_NOT_FOUND", "outcome not found")
	}
	return o, nil
}

func (f *fakeRepo) ListByDecisionIDs(_ context.Context, _ db.Querier, _ []uuid.UUID) ([]Outcome, error) {
	return nil, nil
}

func newTestService(repo Repository, rec recorder, dec Decisions) *service {
	return &service{
		pool:      nil,
		tx:        fakeTx{},
		repo:      repo,
		recorder:  rec,
		decisions: dec,
		clk:       clock.Fixed{T: time.Date(2026, 6, 14, 12, 0, 0, 0, time.UTC)},
	}
}

func ctxWithUser(userID uuid.UUID) context.Context {
	return authctx.WithPrincipal(context.Background(), authctx.Principal{UserID: userID})
}

func TestRecordForbiddenForNonOwnerNonAdmin(t *testing.T) {
	owner := uuid.New()
	other := uuid.New()
	dec := &fakeDecisions{
		decision: decisions.Decision{ID: uuid.New(), OwnerID: owner, Status: decisions.StatusActive},
		role:     teams.RoleMember,
	}
	s := newTestService(newFakeRepo(), &fakeRecorder{}, dec)

	_, err := s.Record(ctxWithUser(other), RecordInput{DecisionID: dec.decision.ID, Summary: "It worked", Success: true})
	if errors.KindOf(err) != errors.KindForbidden || errors.CodeOf(err) != "NOT_DECISION_OWNER" {
		t.Errorf("err = %v, want Forbidden/NOT_DECISION_OWNER", err)
	}
}

func TestRecordRejectsNonDecidableStatus(t *testing.T) {
	owner := uuid.New()
	dec := &fakeDecisions{
		decision: decisions.Decision{ID: uuid.New(), OwnerID: owner, Status: decisions.StatusDraft},
		role:     teams.RoleMember,
	}
	s := newTestService(newFakeRepo(), &fakeRecorder{}, dec)

	_, err := s.Record(ctxWithUser(owner), RecordInput{DecisionID: dec.decision.ID, Summary: "Too early", Success: true})
	if errors.KindOf(err) != errors.KindValidation || errors.CodeOf(err) != "DECISION_NOT_DECIDABLE" {
		t.Errorf("err = %v, want Validation/DECISION_NOT_DECIDABLE", err)
	}
}

func TestRecordOwnerActiveMarksDecidedAndRecordsEvent(t *testing.T) {
	owner := uuid.New()
	dec := &fakeDecisions{
		decision: decisions.Decision{ID: uuid.New(), OwnerID: owner, Status: decisions.StatusActive},
		role:     teams.RoleMember,
	}
	rec := &fakeRecorder{}
	s := newTestService(newFakeRepo(), rec, dec)

	out, err := s.Record(ctxWithUser(owner), RecordInput{DecisionID: dec.decision.ID, Summary: "It worked", Success: true})
	if err != nil {
		t.Fatalf("Record: %v", err)
	}
	if !out.Success {
		t.Errorf("success = false, want true")
	}
	if !out.ResolvedAt.Equal(time.Date(2026, 6, 14, 12, 0, 0, 0, time.UTC)) {
		t.Errorf("ResolvedAt = %v, want fixed clock time", out.ResolvedAt)
	}
	if !dec.markedCalled {
		t.Error("MarkDecided was not called for an active decision")
	}
	if len(rec.events) != 1 || rec.events[0].Type != events.TypeOutcomeRecorded {
		t.Fatalf("events = %+v, want one OutcomeRecorded", rec.events)
	}
	if rec.events[0].AggregateType != aggregateType {
		t.Errorf("aggregate type = %q, want %q", rec.events[0].AggregateType, aggregateType)
	}
}

func TestRecordDecidedDoesNotMarkAgain(t *testing.T) {
	owner := uuid.New()
	dec := &fakeDecisions{
		decision: decisions.Decision{ID: uuid.New(), OwnerID: owner, Status: decisions.StatusDecided},
		role:     teams.RoleMember,
	}
	s := newTestService(newFakeRepo(), &fakeRecorder{}, dec)

	if _, err := s.Record(ctxWithUser(owner), RecordInput{DecisionID: dec.decision.ID, Summary: "Resolved late", Success: false}); err != nil {
		t.Fatalf("Record: %v", err)
	}
	if dec.markedCalled {
		t.Error("MarkDecided was called for an already-decided decision")
	}
}

func TestRecordAdminAllowed(t *testing.T) {
	owner := uuid.New()
	admin := uuid.New()
	dec := &fakeDecisions{
		decision: decisions.Decision{ID: uuid.New(), OwnerID: owner, Status: decisions.StatusActive},
		role:     teams.RoleAdmin,
	}
	s := newTestService(newFakeRepo(), &fakeRecorder{}, dec)

	if _, err := s.Record(ctxWithUser(admin), RecordInput{DecisionID: dec.decision.ID, Summary: "Admin closed it", Success: true}); err != nil {
		t.Fatalf("Record by admin: %v", err)
	}
}

func TestRecordTwiceUpserts(t *testing.T) {
	owner := uuid.New()
	repo := newFakeRepo()
	dec := &fakeDecisions{
		decision: decisions.Decision{ID: uuid.New(), OwnerID: owner, Status: decisions.StatusActive},
		role:     teams.RoleMember,
	}
	s := newTestService(repo, &fakeRecorder{}, dec)

	first, err := s.Record(ctxWithUser(owner), RecordInput{DecisionID: dec.decision.ID, Summary: "First", Success: true})
	if err != nil {
		t.Fatalf("Record first: %v", err)
	}
	second, err := s.Record(ctxWithUser(owner), RecordInput{DecisionID: dec.decision.ID, Summary: "Revised", Success: false})
	if err != nil {
		t.Fatalf("Record second: %v", err)
	}
	if first.ID != second.ID {
		t.Errorf("upsert produced new id: %v != %v", first.ID, second.ID)
	}
	if len(repo.byDecision) != 1 {
		t.Errorf("repo holds %d outcomes, want 1 (upsert)", len(repo.byDecision))
	}
	if second.Summary != "Revised" || second.Success {
		t.Errorf("second = %+v, want Revised/false", second)
	}
}

func (f *fakeRepo) ListByIDs(_ context.Context, _ db.Querier, _ []uuid.UUID) ([]Outcome, error) {
	return nil, nil
}
