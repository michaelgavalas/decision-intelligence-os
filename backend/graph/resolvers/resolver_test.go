package resolvers_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/michaelgavalas/decision-intelligence-os/backend/graph/model"
	"github.com/michaelgavalas/decision-intelligence-os/backend/graph/resolvers"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/analytics"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/decisions"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/db"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/users"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/authctx"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/errors"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/pagination"
)

// authedCtx returns a context carrying an arbitrary authenticated principal.
func authedCtx() context.Context {
	return authctx.WithPrincipal(context.Background(), authctx.Principal{UserID: uuid.New()})
}

// fakeDecisions is a minimal decisions.Service double; only Create is exercised.
type fakeDecisions struct {
	created decisions.Decision
	err     error
}

func (f *fakeDecisions) Create(_ context.Context, in decisions.CreateInput) (decisions.Decision, error) {
	if f.err != nil {
		return decisions.Decision{}, f.err
	}
	d := f.created
	d.TeamID = in.TeamID
	d.Title = in.Title
	return d, nil
}

func (f *fakeDecisions) GetByID(context.Context, uuid.UUID) (decisions.Decision, error) {
	return decisions.Decision{}, nil
}

func (f *fakeDecisions) AuthorizeAccess(context.Context, uuid.UUID) (decisions.Decision, string, error) {
	return decisions.Decision{}, "", nil
}

func (f *fakeDecisions) List(context.Context, uuid.UUID, pagination.PageArgs) (pagination.Connection[decisions.Decision], error) {
	return pagination.Connection[decisions.Decision]{}, nil
}

func (f *fakeDecisions) Update(context.Context, uuid.UUID, decisions.UpdateInput) (decisions.Decision, error) {
	return decisions.Decision{}, nil
}

func (f *fakeDecisions) Transition(context.Context, uuid.UUID, string) (decisions.Decision, error) {
	return decisions.Decision{}, nil
}

func (f *fakeDecisions) MarkDecided(context.Context, db.Querier, uuid.UUID) (decisions.Decision, error) {
	return decisions.Decision{}, nil
}

// fakeUsers is a minimal users.Service double; only GetByID is exercised.
type fakeUsers struct {
	user users.User
	err  error
}

func (f *fakeUsers) Provision(context.Context, db.Querier, users.ProvisionParams) (users.User, error) {
	return users.User{}, nil
}

func (f *fakeUsers) GetByID(context.Context, uuid.UUID) (users.User, error) {
	if f.err != nil {
		return users.User{}, f.err
	}
	return f.user, nil
}

func (f *fakeUsers) FindByEmail(context.Context, string) (users.User, error) {
	return users.User{}, nil
}

func (f *fakeUsers) UpdateProfile(context.Context, uuid.UUID, string) (users.User, error) {
	return users.User{}, nil
}

// fakeAnalytics is a minimal analytics.Service double.
type fakeAnalytics struct {
	metrics analytics.TeamMetrics
	err     error
}

func (f *fakeAnalytics) TeamMetrics(context.Context, uuid.UUID) (analytics.TeamMetrics, error) {
	if f.err != nil {
		return analytics.TeamMetrics{}, f.err
	}
	return f.metrics, nil
}

func (f *fakeAnalytics) Calibration(context.Context, uuid.UUID) (analytics.CalibrationReport, error) {
	return analytics.CalibrationReport{}, nil
}

func TestCreateDecisionSuccess(t *testing.T) {
	teamID := uuid.New()
	decisionID := uuid.New()
	r := &resolvers.Resolver{
		Decisions: &fakeDecisions{created: decisions.Decision{ID: decisionID, OwnerID: uuid.New(), Status: decisions.StatusDraft}},
	}

	payload, err := r.Mutation().CreateDecision(authedCtx(), model.CreateDecisionInput{
		TeamID: teamID.String(),
		Title:  "Launch",
	})
	if err != nil {
		t.Fatalf("CreateDecision: unexpected transport error: %v", err)
	}
	if payload.Decision == nil {
		t.Fatal("CreateDecision: expected a decision in the payload")
	}
	if payload.Decision.ID != decisionID.String() {
		t.Errorf("decision id = %q, want %q", payload.Decision.ID, decisionID.String())
	}
	if payload.Decision.Title != "Launch" {
		t.Errorf("decision title = %q, want Launch", payload.Decision.Title)
	}
	if len(payload.UserErrors) != 0 {
		t.Errorf("userErrors = %v, want empty", payload.UserErrors)
	}
}

func TestCreateDecisionValidationBecomesUserError(t *testing.T) {
	r := &resolvers.Resolver{
		Decisions: &fakeDecisions{err: errors.Validation("DECISION_TITLE_REQUIRED", "decision title is required")},
	}

	payload, err := r.Mutation().CreateDecision(authedCtx(), model.CreateDecisionInput{
		TeamID: uuid.New().String(),
	})
	if err != nil {
		t.Fatalf("CreateDecision: expected nil transport error, got %v", err)
	}
	if payload.Decision != nil {
		t.Errorf("expected no decision, got %+v", payload.Decision)
	}
	if len(payload.UserErrors) != 1 {
		t.Fatalf("userErrors len = %d, want 1", len(payload.UserErrors))
	}
	if payload.UserErrors[0].Code != "DECISION_TITLE_REQUIRED" {
		t.Errorf("userError code = %q, want DECISION_TITLE_REQUIRED", payload.UserErrors[0].Code)
	}
}

func TestCreateDecisionInternalBecomesTransportError(t *testing.T) {
	r := &resolvers.Resolver{
		Decisions: &fakeDecisions{err: errors.Internal("BOOM", "something failed")},
	}

	payload, err := r.Mutation().CreateDecision(authedCtx(), model.CreateDecisionInput{
		TeamID: uuid.New().String(),
	})
	if err == nil {
		t.Fatal("CreateDecision: expected a transport error for an internal failure")
	}
	if payload != nil {
		t.Errorf("expected nil payload, got %+v", payload)
	}
}

func TestMeReturnsMappedUser(t *testing.T) {
	userID := uuid.New()
	r := &resolvers.Resolver{
		Users: &fakeUsers{user: users.User{ID: userID, Email: "ada@example.com", Name: "Ada"}},
	}

	user, err := r.Query().Me(authedCtx())
	if err != nil {
		t.Fatalf("Me: %v", err)
	}
	if user.ID != userID.String() || user.Email != "ada@example.com" || user.Name != "Ada" {
		t.Errorf("Me = %+v, want mapped user", user)
	}
}

func TestTeamMetricsReturnsMapped(t *testing.T) {
	r := &resolvers.Resolver{
		Analytics: &fakeAnalytics{metrics: analytics.TeamMetrics{
			BrierScore:            0.2,
			ForecastCount:         5,
			DecisionSuccessRate:   0.6,
			ResolvedDecisionCount: 3,
		}},
	}

	metrics, err := r.Query().TeamMetrics(authedCtx(), uuid.New().String())
	if err != nil {
		t.Fatalf("TeamMetrics: %v", err)
	}
	if metrics.BrierScore != 0.2 || metrics.ForecastCount != 5 || metrics.DecisionSuccessRate != 0.6 || metrics.ResolvedDecisionCount != 3 {
		t.Errorf("TeamMetrics = %+v, want mapped values", metrics)
	}
}
