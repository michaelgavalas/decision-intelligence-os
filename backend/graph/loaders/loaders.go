// Package loaders provides per-request DataLoaders that batch and cache the
// repository reads the GraphQL resolvers perform while assembling nested object
// trees. Without batching, resolving (for example) the owner of every decision
// in a list would issue one query per decision; the loaders coalesce those into
// a single ListByIDs call per relation per request.
//
// Loaders MUST be created per request: their cache is scoped to a single
// GraphQL operation so reads never leak across requests. Middleware installs a
// fresh *Loaders into each request's context, and the accessor helpers read it
// back out.
package loaders

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vikstrous/dataloadgen"

	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/assumptions"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/decisions"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/evidence"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/outcomes"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/predictions"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/teams"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/users"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/errors"
)

// Loaders holds one batching loader per relation the resolvers traverse. The
// by-id loaders return a single entity per key; the by-parent loaders return a
// slice per key (the children grouped under that parent).
type Loaders struct {
	User       *dataloadgen.Loader[uuid.UUID, users.User]
	Team       *dataloadgen.Loader[uuid.UUID, teams.Team]
	Decision   *dataloadgen.Loader[uuid.UUID, decisions.Decision]
	Assumption *dataloadgen.Loader[uuid.UUID, assumptions.Assumption]
	Evidence   *dataloadgen.Loader[uuid.UUID, evidence.Evidence]
	Prediction *dataloadgen.Loader[uuid.UUID, predictions.Prediction]
	Outcome    *dataloadgen.Loader[uuid.UUID, outcomes.Outcome]

	AssumptionsByDecision *dataloadgen.Loader[uuid.UUID, []assumptions.Assumption]
	EvidenceByAssumption  *dataloadgen.Loader[uuid.UUID, []evidence.Evidence]
	PredictionsByDecision *dataloadgen.Loader[uuid.UUID, []predictions.Prediction]
	OutcomeByDecision     *dataloadgen.Loader[uuid.UUID, *outcomes.Outcome]
	MembersByTeam         *dataloadgen.Loader[uuid.UUID, []teams.Membership]
}

// NewLoaders wires a fresh set of per-request loaders backed by the supplied
// repositories and pool. All reads run through the pool (loaders are read-only
// and never join a write transaction).
func NewLoaders(
	pool *pgxpool.Pool,
	usersRepo users.Repository,
	teamsRepo teams.Repository,
	decisionsRepo decisions.Repository,
	assumptionsRepo assumptions.Repository,
	evidenceRepo evidence.Repository,
	predictionsRepo predictions.Repository,
	outcomesRepo outcomes.Repository,
) *Loaders {
	return &Loaders{
		User: dataloadgen.NewLoader(func(ctx context.Context, keys []uuid.UUID) ([]users.User, []error) {
			rows, err := usersRepo.ListByIDs(ctx, pool, keys)
			if err != nil {
				return errResults[users.User](keys, err)
			}
			byID := make(map[uuid.UUID]users.User, len(rows))
			for _, r := range rows {
				byID[r.ID] = r
			}
			return mapByKey(keys, byID, "USER_NOT_FOUND", "user not found")
		}),
		Team: dataloadgen.NewLoader(func(ctx context.Context, keys []uuid.UUID) ([]teams.Team, []error) {
			rows, err := teamsRepo.ListByIDs(ctx, pool, keys)
			if err != nil {
				return errResults[teams.Team](keys, err)
			}
			byID := make(map[uuid.UUID]teams.Team, len(rows))
			for _, r := range rows {
				byID[r.ID] = r
			}
			return mapByKey(keys, byID, "TEAM_NOT_FOUND", "team not found")
		}),
		Decision: dataloadgen.NewLoader(func(ctx context.Context, keys []uuid.UUID) ([]decisions.Decision, []error) {
			rows, err := decisionsRepo.ListByIDs(ctx, pool, keys)
			if err != nil {
				return errResults[decisions.Decision](keys, err)
			}
			byID := make(map[uuid.UUID]decisions.Decision, len(rows))
			for _, r := range rows {
				byID[r.ID] = r
			}
			return mapByKey(keys, byID, "DECISION_NOT_FOUND", "decision not found")
		}),
		Assumption: dataloadgen.NewLoader(func(ctx context.Context, keys []uuid.UUID) ([]assumptions.Assumption, []error) {
			rows, err := assumptionsRepo.ListByIDs(ctx, pool, keys)
			if err != nil {
				return errResults[assumptions.Assumption](keys, err)
			}
			byID := make(map[uuid.UUID]assumptions.Assumption, len(rows))
			for _, r := range rows {
				byID[r.ID] = r
			}
			return mapByKey(keys, byID, "ASSUMPTION_NOT_FOUND", "assumption not found")
		}),
		Evidence: dataloadgen.NewLoader(func(ctx context.Context, keys []uuid.UUID) ([]evidence.Evidence, []error) {
			rows, err := evidenceRepo.ListByIDs(ctx, pool, keys)
			if err != nil {
				return errResults[evidence.Evidence](keys, err)
			}
			byID := make(map[uuid.UUID]evidence.Evidence, len(rows))
			for _, r := range rows {
				byID[r.ID] = r
			}
			return mapByKey(keys, byID, "EVIDENCE_NOT_FOUND", "evidence not found")
		}),
		Prediction: dataloadgen.NewLoader(func(ctx context.Context, keys []uuid.UUID) ([]predictions.Prediction, []error) {
			rows, err := predictionsRepo.ListByIDs(ctx, pool, keys)
			if err != nil {
				return errResults[predictions.Prediction](keys, err)
			}
			byID := make(map[uuid.UUID]predictions.Prediction, len(rows))
			for _, r := range rows {
				byID[r.ID] = r
			}
			return mapByKey(keys, byID, "PREDICTION_NOT_FOUND", "prediction not found")
		}),
		Outcome: dataloadgen.NewLoader(func(ctx context.Context, keys []uuid.UUID) ([]outcomes.Outcome, []error) {
			rows, err := outcomesRepo.ListByIDs(ctx, pool, keys)
			if err != nil {
				return errResults[outcomes.Outcome](keys, err)
			}
			byID := make(map[uuid.UUID]outcomes.Outcome, len(rows))
			for _, r := range rows {
				byID[r.ID] = r
			}
			return mapByKey(keys, byID, "OUTCOME_NOT_FOUND", "outcome not found")
		}),

		AssumptionsByDecision: dataloadgen.NewLoader(func(ctx context.Context, keys []uuid.UUID) ([][]assumptions.Assumption, []error) {
			rows, err := assumptionsRepo.ListByDecisionIDs(ctx, pool, keys)
			if err != nil {
				return errResults[[]assumptions.Assumption](keys, err)
			}
			grouped := make(map[uuid.UUID][]assumptions.Assumption, len(keys))
			for _, r := range rows {
				grouped[r.DecisionID] = append(grouped[r.DecisionID], r)
			}
			return groupByKey(keys, grouped), nil
		}),
		EvidenceByAssumption: dataloadgen.NewLoader(func(ctx context.Context, keys []uuid.UUID) ([][]evidence.Evidence, []error) {
			rows, err := evidenceRepo.ListByAssumptionIDs(ctx, pool, keys)
			if err != nil {
				return errResults[[]evidence.Evidence](keys, err)
			}
			grouped := make(map[uuid.UUID][]evidence.Evidence, len(keys))
			for _, r := range rows {
				grouped[r.AssumptionID] = append(grouped[r.AssumptionID], r)
			}
			return groupByKey(keys, grouped), nil
		}),
		PredictionsByDecision: dataloadgen.NewLoader(func(ctx context.Context, keys []uuid.UUID) ([][]predictions.Prediction, []error) {
			rows, err := predictionsRepo.ListByDecisionIDs(ctx, pool, keys)
			if err != nil {
				return errResults[[]predictions.Prediction](keys, err)
			}
			grouped := make(map[uuid.UUID][]predictions.Prediction, len(keys))
			for _, r := range rows {
				grouped[r.DecisionID] = append(grouped[r.DecisionID], r)
			}
			return groupByKey(keys, grouped), nil
		}),
		OutcomeByDecision: dataloadgen.NewLoader(func(ctx context.Context, keys []uuid.UUID) ([]*outcomes.Outcome, []error) {
			rows, err := outcomesRepo.ListByDecisionIDs(ctx, pool, keys)
			if err != nil {
				return errResults[*outcomes.Outcome](keys, err)
			}
			byDecision := make(map[uuid.UUID]outcomes.Outcome, len(rows))
			for _, r := range rows {
				byDecision[r.DecisionID] = r
			}
			out := make([]*outcomes.Outcome, len(keys))
			for i, k := range keys {
				if o, ok := byDecision[k]; ok {
					oc := o
					out[i] = &oc
				}
			}
			return out, nil
		}),
		MembersByTeam: dataloadgen.NewLoader(func(ctx context.Context, keys []uuid.UUID) ([][]teams.Membership, []error) {
			rows, err := teamsRepo.ListMembersByTeamIDs(ctx, pool, keys)
			if err != nil {
				return errResults[[]teams.Membership](keys, err)
			}
			grouped := make(map[uuid.UUID][]teams.Membership, len(keys))
			for _, r := range rows {
				grouped[r.TeamID] = append(grouped[r.TeamID], r)
			}
			return groupByKey(keys, grouped), nil
		}),
	}
}

// mapByKey returns one value per key in order, pulling each from byID and
// substituting a NotFound error for any key the batch did not return.
func mapByKey[V any](keys []uuid.UUID, byID map[uuid.UUID]V, code, msg string) ([]V, []error) {
	values := make([]V, len(keys))
	var errs []error
	for i, k := range keys {
		v, ok := byID[k]
		if !ok {
			if errs == nil {
				errs = make([]error, len(keys))
			}
			errs[i] = errors.NotFound(code, msg)
			continue
		}
		values[i] = v
	}
	return values, errs
}

// groupByKey returns one slice per key in order; a key with no children yields
// an empty (non-nil) slice. A missing parent is not an error for to-many
// relations.
func groupByKey[V any](keys []uuid.UUID, grouped map[uuid.UUID][]V) [][]V {
	out := make([][]V, len(keys))
	for i, k := range keys {
		if children, ok := grouped[k]; ok {
			out[i] = children
		} else {
			out[i] = []V{}
		}
	}
	return out
}

// errResults builds a same-length value and error slice that fails every key
// with err, used when the batch query itself fails.
func errResults[V any](keys []uuid.UUID, err error) ([]V, []error) {
	values := make([]V, len(keys))
	errs := make([]error, len(keys))
	for i := range keys {
		errs[i] = err
	}
	return values, errs
}

type contextKey struct{}

// loadersKey is the unexported context key under which the per-request *Loaders
// is stored.
var loadersKey = contextKey{}

// Middleware returns HTTP middleware that builds a fresh *Loaders for every
// request and stores it in the request context, so loader caches never leak
// across requests.
func Middleware(
	pool *pgxpool.Pool,
	usersRepo users.Repository,
	teamsRepo teams.Repository,
	decisionsRepo decisions.Repository,
	assumptionsRepo assumptions.Repository,
	evidenceRepo evidence.Repository,
	predictionsRepo predictions.Repository,
	outcomesRepo outcomes.Repository,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			l := NewLoaders(pool, usersRepo, teamsRepo, decisionsRepo, assumptionsRepo, evidenceRepo, predictionsRepo, outcomesRepo)
			ctx := context.WithValue(r.Context(), loadersKey, l)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// WithLoaders returns a copy of ctx carrying loaders. It exists so callers
// outside the HTTP middleware (such as subscription resolvers driven by a
// background context) can install a loader set explicitly.
func WithLoaders(ctx context.Context, l *Loaders) context.Context {
	return context.WithValue(ctx, loadersKey, l)
}

// For returns the *Loaders stored in ctx, or nil when none is present.
func For(ctx context.Context) *Loaders {
	l, _ := ctx.Value(loadersKey).(*Loaders)
	return l
}

// errNoLoaders is returned by the accessors when no per-request loader set is
// present in the context, which indicates a wiring bug rather than a client
// error.
func errNoLoaders() error {
	return errors.Internal("LOADERS_MISSING", "data loaders are not present in the request context")
}

// GetUser loads a single user by id through the per-request loader.
func GetUser(ctx context.Context, id uuid.UUID) (users.User, error) {
	l := For(ctx)
	if l == nil {
		return users.User{}, errNoLoaders()
	}
	return l.User.Load(ctx, id)
}

// GetTeam loads a single team by id through the per-request loader.
func GetTeam(ctx context.Context, id uuid.UUID) (teams.Team, error) {
	l := For(ctx)
	if l == nil {
		return teams.Team{}, errNoLoaders()
	}
	return l.Team.Load(ctx, id)
}

// GetDecision loads a single decision by id through the per-request loader.
func GetDecision(ctx context.Context, id uuid.UUID) (decisions.Decision, error) {
	l := For(ctx)
	if l == nil {
		return decisions.Decision{}, errNoLoaders()
	}
	return l.Decision.Load(ctx, id)
}

// GetAssumption loads a single assumption by id through the per-request loader.
func GetAssumption(ctx context.Context, id uuid.UUID) (assumptions.Assumption, error) {
	l := For(ctx)
	if l == nil {
		return assumptions.Assumption{}, errNoLoaders()
	}
	return l.Assumption.Load(ctx, id)
}

// GetEvidence loads a single piece of evidence by id through the per-request
// loader.
func GetEvidence(ctx context.Context, id uuid.UUID) (evidence.Evidence, error) {
	l := For(ctx)
	if l == nil {
		return evidence.Evidence{}, errNoLoaders()
	}
	return l.Evidence.Load(ctx, id)
}

// GetPrediction loads a single prediction by id through the per-request loader.
func GetPrediction(ctx context.Context, id uuid.UUID) (predictions.Prediction, error) {
	l := For(ctx)
	if l == nil {
		return predictions.Prediction{}, errNoLoaders()
	}
	return l.Prediction.Load(ctx, id)
}

// GetOutcome loads a single outcome by id through the per-request loader.
func GetOutcome(ctx context.Context, id uuid.UUID) (outcomes.Outcome, error) {
	l := For(ctx)
	if l == nil {
		return outcomes.Outcome{}, errNoLoaders()
	}
	return l.Outcome.Load(ctx, id)
}

// AssumptionsByDecision loads a decision's assumptions through the per-request
// loader.
func AssumptionsByDecision(ctx context.Context, decisionID uuid.UUID) ([]assumptions.Assumption, error) {
	l := For(ctx)
	if l == nil {
		return nil, errNoLoaders()
	}
	return l.AssumptionsByDecision.Load(ctx, decisionID)
}

// EvidenceByAssumption loads an assumption's evidence through the per-request
// loader.
func EvidenceByAssumption(ctx context.Context, assumptionID uuid.UUID) ([]evidence.Evidence, error) {
	l := For(ctx)
	if l == nil {
		return nil, errNoLoaders()
	}
	return l.EvidenceByAssumption.Load(ctx, assumptionID)
}

// PredictionsByDecision loads a decision's predictions through the per-request
// loader.
func PredictionsByDecision(ctx context.Context, decisionID uuid.UUID) ([]predictions.Prediction, error) {
	l := For(ctx)
	if l == nil {
		return nil, errNoLoaders()
	}
	return l.PredictionsByDecision.Load(ctx, decisionID)
}

// OutcomeByDecision loads a decision's outcome through the per-request loader,
// returning nil when the decision has no recorded outcome.
func OutcomeByDecision(ctx context.Context, decisionID uuid.UUID) (*outcomes.Outcome, error) {
	l := For(ctx)
	if l == nil {
		return nil, errNoLoaders()
	}
	return l.OutcomeByDecision.Load(ctx, decisionID)
}

// MembersByTeam loads a team's memberships through the per-request loader.
func MembersByTeam(ctx context.Context, teamID uuid.UUID) ([]teams.Membership, error) {
	l := For(ctx)
	if l == nil {
		return nil, errNoLoaders()
	}
	return l.MembersByTeam.Load(ctx, teamID)
}
