package evidence

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/assumptions"
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

// fakeAssumptions is an in-memory Assumptions dependency.
type fakeAssumptions struct {
	role      string
	notMember bool
}

func (f *fakeAssumptions) AuthorizeAccess(_ context.Context, id uuid.UUID) (assumptions.Assumption, string, error) {
	if f.notMember {
		return assumptions.Assumption{}, "", errors.Forbidden("NOT_TEAM_MEMBER", "not a member")
	}
	return assumptions.Assumption{ID: id}, f.role, nil
}

// fakeRepo is an in-memory Repository.
type fakeRepo struct {
	byID map[uuid.UUID]Evidence
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{byID: map[uuid.UUID]Evidence{}}
}

func (f *fakeRepo) Create(_ context.Context, _ db.Querier, e Evidence) (Evidence, error) {
	e.CreatedAt = time.Now()
	e.UpdatedAt = e.CreatedAt
	f.byID[e.ID] = e
	return e, nil
}

func (f *fakeRepo) GetByID(_ context.Context, _ db.Querier, id uuid.UUID) (Evidence, error) {
	e, ok := f.byID[id]
	if !ok {
		return Evidence{}, errors.NotFound("EVIDENCE_NOT_FOUND", "evidence not found")
	}
	return e, nil
}

func (f *fakeRepo) ListByAssumption(_ context.Context, _ db.Querier, _ uuid.UUID) ([]Evidence, error) {
	return nil, nil
}

func (f *fakeRepo) ListByAssumptionIDs(_ context.Context, _ db.Querier, _ []uuid.UUID) ([]Evidence, error) {
	return nil, nil
}

func (f *fakeRepo) Update(_ context.Context, _ db.Querier, id uuid.UUID, sourceType string, sourceURL *string, content string) (Evidence, error) {
	e, ok := f.byID[id]
	if !ok {
		return Evidence{}, errors.NotFound("EVIDENCE_NOT_FOUND", "evidence not found")
	}
	e.SourceType = sourceType
	e.SourceURL = sourceURL
	e.Content = content
	f.byID[id] = e
	return e, nil
}

func (f *fakeRepo) Delete(_ context.Context, _ db.Querier, id uuid.UUID) error {
	delete(f.byID, id)
	return nil
}

func newTestService(repo Repository, rec recorder, as Assumptions) *service {
	return &service{pool: nil, tx: fakeTx{}, repo: repo, recorder: rec, assumptions: as}
}

func ctxWithUser(userID uuid.UUID) context.Context {
	return authctx.WithPrincipal(context.Background(), authctx.Principal{UserID: userID})
}

func TestAttachInvalidSourceType(t *testing.T) {
	user := uuid.New()
	as := &fakeAssumptions{role: teams.RoleMember}
	s := newTestService(newFakeRepo(), &fakeRecorder{}, as)

	_, err := s.Attach(ctxWithUser(user), AttachInput{AssumptionID: uuid.New(), SourceType: "bogus", Content: "x"})
	if errors.KindOf(err) != errors.KindValidation || errors.CodeOf(err) != "INVALID_SOURCE_TYPE" {
		t.Errorf("err = %v, want Validation/INVALID_SOURCE_TYPE", err)
	}
}

func TestAttachForbiddenForViewer(t *testing.T) {
	user := uuid.New()
	as := &fakeAssumptions{role: teams.RoleViewer}
	s := newTestService(newFakeRepo(), &fakeRecorder{}, as)

	_, err := s.Attach(ctxWithUser(user), AttachInput{AssumptionID: uuid.New(), SourceType: SourceNote, Content: "x"})
	if errors.KindOf(err) != errors.KindForbidden || errors.CodeOf(err) != "EVIDENCE_FORBIDDEN" {
		t.Errorf("err = %v, want Forbidden/EVIDENCE_FORBIDDEN", err)
	}
}

func TestAttachForbiddenForNonMember(t *testing.T) {
	user := uuid.New()
	as := &fakeAssumptions{notMember: true}
	s := newTestService(newFakeRepo(), &fakeRecorder{}, as)

	_, err := s.Attach(ctxWithUser(user), AttachInput{AssumptionID: uuid.New(), SourceType: SourceNote, Content: "x"})
	if errors.KindOf(err) != errors.KindForbidden {
		t.Errorf("err = %v, want Forbidden", err)
	}
}

func TestAttachSucceedsAndRecordsEvent(t *testing.T) {
	user := uuid.New()
	as := &fakeAssumptions{role: teams.RoleMember}
	rec := &fakeRecorder{}
	s := newTestService(newFakeRepo(), rec, as)

	url := "https://example.com/report"
	ev, err := s.Attach(ctxWithUser(user), AttachInput{
		AssumptionID: uuid.New(), SourceType: SourceURL, SourceURL: &url, Content: "Supporting data",
	})
	if err != nil {
		t.Fatalf("Attach: %v", err)
	}
	if ev.SourceType != SourceURL || ev.SourceURL == nil || *ev.SourceURL != url {
		t.Errorf("evidence = %+v, want url source with link", ev)
	}
	if len(rec.events) != 1 || rec.events[0].Type != events.TypeEvidenceAttached {
		t.Fatalf("events = %+v, want one EvidenceAttached", rec.events)
	}
	if rec.events[0].ActorID == nil || *rec.events[0].ActorID != user {
		t.Errorf("actor = %v, want %v", rec.events[0].ActorID, user)
	}
}

func TestRemoveForbiddenForViewer(t *testing.T) {
	user := uuid.New()
	repo := newFakeRepo()
	ev, _ := repo.Create(context.Background(), nil, Evidence{ID: uuid.New(), AssumptionID: uuid.New(), SourceType: SourceNote})
	as := &fakeAssumptions{role: teams.RoleViewer}
	s := newTestService(repo, &fakeRecorder{}, as)

	err := s.Remove(ctxWithUser(user), ev.ID)
	if errors.KindOf(err) != errors.KindForbidden {
		t.Errorf("err = %v, want Forbidden", err)
	}
}

func (f *fakeRepo) ListByIDs(_ context.Context, _ db.Querier, _ []uuid.UUID) ([]Evidence, error) {
	return nil, nil
}
