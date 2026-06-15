package resolvers

import (
	"context"

	"github.com/google/uuid"

	"github.com/michaelgavalas/decision-intelligence-os/backend/graph/loaders"
	"github.com/michaelgavalas/decision-intelligence-os/backend/graph/model"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/ai"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/analytics"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/assumptions"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/decisions"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/evidence"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/outcomes"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/predictions"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/teams"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/users"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/errors"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/pagination"
)

// The mappers translate domain models into the GraphQL model types. They set
// only the scalar and id fields; every nested object field (Decision.owner,
// Assumption.evidence, ...) is populated lazily by a field resolver backed by a
// dataloader, so the mappers never traverse the object graph.

// toUser maps a domain user to its GraphQL model.
func toUser(u users.User) *model.User {
	return &model.User{
		ID:        u.ID.String(),
		Email:     u.Email,
		Name:      u.Name,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// toTeam maps a domain team to its GraphQL model.
func toTeam(t teams.Team) *model.Team {
	return &model.Team{
		ID:        t.ID.String(),
		Name:      t.Name,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}

// toMembership maps a domain membership to its GraphQL model. Unlike the other
// types, a membership exposes no id of its own, so its team and user objects
// cannot be resolved later from a parent id; they are resolved eagerly here
// through the per-request loaders (which still batch the lookups).
func toMembership(ctx context.Context, m teams.Membership) (*model.Membership, error) {
	team, err := loaders.GetTeam(ctx, m.TeamID)
	if err != nil {
		return nil, err
	}
	user, err := loaders.GetUser(ctx, m.UserID)
	if err != nil {
		return nil, err
	}
	return &model.Membership{
		Team:      toTeam(team),
		User:      toUser(user),
		Role:      roleToModel(m.Role),
		CreatedAt: m.CreatedAt,
	}, nil
}

// toDecision maps a domain decision to its GraphQL model.
func toDecision(d decisions.Decision) *model.Decision {
	return &model.Decision{
		ID:          d.ID.String(),
		Title:       d.Title,
		Description: d.Description,
		Status:      decisionStatusToModel(d.Status),
		DecidedAt:   d.DecidedAt,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
}

// toAssumption maps a domain assumption to its GraphQL model.
func toAssumption(a assumptions.Assumption) *model.Assumption {
	return &model.Assumption{
		ID:         a.ID.String(),
		Statement:  a.Statement,
		Confidence: a.Confidence,
		CreatedAt:  a.CreatedAt,
		UpdatedAt:  a.UpdatedAt,
	}
}

// toEvidence maps a domain evidence record to its GraphQL model.
func toEvidence(e evidence.Evidence) *model.Evidence {
	return &model.Evidence{
		ID:         e.ID.String(),
		SourceType: evidenceSourceTypeToModel(e.SourceType),
		SourceURL:  e.SourceURL,
		Content:    e.Content,
		CreatedAt:  e.CreatedAt,
		UpdatedAt:  e.UpdatedAt,
	}
}

// toPrediction maps a domain prediction to its GraphQL model.
func toPrediction(p predictions.Prediction) *model.Prediction {
	return &model.Prediction{
		ID:          p.ID.String(),
		Statement:   p.Statement,
		Probability: p.Probability,
		ResolvesAt:  p.ResolvesAt,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

// toOutcome maps a domain outcome to its GraphQL model.
func toOutcome(o outcomes.Outcome) *model.Outcome {
	return &model.Outcome{
		ID:         o.ID.String(),
		Summary:    o.Summary,
		Success:    o.Success,
		ResolvedAt: o.ResolvedAt,
		CreatedAt:  o.CreatedAt,
		UpdatedAt:  o.UpdatedAt,
	}
}

// toDecisionConnection maps a domain pagination connection of decisions onto
// the GraphQL connection type.
func toDecisionConnection(conn pagination.Connection[decisions.Decision]) *model.DecisionConnection {
	edges := make([]model.DecisionEdge, 0, len(conn.Edges))
	for _, e := range conn.Edges {
		edges = append(edges, model.DecisionEdge{
			Node:   toDecision(e.Node),
			Cursor: e.Cursor,
		})
	}
	return &model.DecisionConnection{
		Edges: edges,
		PageInfo: &model.PageInfo{
			HasNextPage:     conn.PageInfo.HasNextPage,
			HasPreviousPage: conn.PageInfo.HasPreviousPage,
			StartCursor:     conn.PageInfo.StartCursor,
			EndCursor:       conn.PageInfo.EndCursor,
		},
		TotalCount: conn.TotalCount,
	}
}

// toTeamMetrics maps a domain metrics summary to its GraphQL model.
func toTeamMetrics(m analytics.TeamMetrics) *model.TeamMetrics {
	return &model.TeamMetrics{
		BrierScore:            m.BrierScore,
		ForecastCount:         m.ForecastCount,
		DecisionSuccessRate:   m.DecisionSuccessRate,
		ResolvedDecisionCount: m.ResolvedDecisionCount,
	}
}

// toCalibrationReport maps a domain calibration report to its GraphQL model.
func toCalibrationReport(r analytics.CalibrationReport) *model.CalibrationReport {
	bins := make([]model.CalibrationBin, 0, len(r.Bins))
	for _, b := range r.Bins {
		bins = append(bins, model.CalibrationBin{
			Bucket:            b.Bucket,
			MeanPredicted:     b.MeanPredicted,
			ObservedFrequency: b.ObservedFrequency,
			SampleSize:        b.SampleSize,
		})
	}
	return &model.CalibrationReport{Bins: bins}
}

// toBiasReport maps a domain bias report to its GraphQL model.
func toBiasReport(r ai.BiasReport) *model.BiasReport {
	biases := make([]model.DetectedBias, 0, len(r.Biases))
	for _, b := range r.Biases {
		biases = append(biases, model.DetectedBias{Name: b.Name, Explanation: b.Explanation})
	}
	return &model.BiasReport{Summary: r.Summary, Biases: biases}
}

// parseID parses a GraphQL ID string into a UUID, returning a validation error
// the transport layer maps to a client-facing message.
func parseID(s string) (uuid.UUID, error) {
	parsed, err := uuid.Parse(s)
	if err != nil {
		return uuid.UUID{}, errors.Validation("INVALID_ID", "id is not a valid identifier")
	}
	return parsed, nil
}

// roleToModel maps a domain role string to the GraphQL enum.
func roleToModel(role string) model.Role {
	switch role {
	case teams.RoleAdmin:
		return model.RoleAdmin
	case teams.RoleMember:
		return model.RoleMember
	default:
		return model.RoleViewer
	}
}

// roleFromModel maps a GraphQL role enum to the domain role string.
func roleFromModel(role model.Role) string {
	switch role {
	case model.RoleAdmin:
		return teams.RoleAdmin
	case model.RoleMember:
		return teams.RoleMember
	default:
		return teams.RoleViewer
	}
}

// decisionStatusToModel maps a domain status string to the GraphQL enum.
func decisionStatusToModel(status string) model.DecisionStatus {
	switch status {
	case decisions.StatusDraft:
		return model.DecisionStatusDraft
	case decisions.StatusActive:
		return model.DecisionStatusActive
	case decisions.StatusDecided:
		return model.DecisionStatusDecided
	default:
		return model.DecisionStatusArchived
	}
}

// decisionStatusFromModel maps a GraphQL status enum to the domain status
// string.
func decisionStatusFromModel(status model.DecisionStatus) string {
	switch status {
	case model.DecisionStatusDraft:
		return decisions.StatusDraft
	case model.DecisionStatusActive:
		return decisions.StatusActive
	case model.DecisionStatusDecided:
		return decisions.StatusDecided
	default:
		return decisions.StatusArchived
	}
}

// evidenceSourceTypeToModel maps a domain source type string to the GraphQL
// enum.
func evidenceSourceTypeToModel(sourceType string) model.EvidenceSourceType {
	switch sourceType {
	case evidence.SourceURL:
		return model.EvidenceSourceTypeURL
	case evidence.SourceDocument:
		return model.EvidenceSourceTypeDocument
	case evidence.SourceNote:
		return model.EvidenceSourceTypeNote
	default:
		return model.EvidenceSourceTypeDataset
	}
}

// evidenceSourceTypeFromModel maps a GraphQL source-type enum to the domain
// string.
func evidenceSourceTypeFromModel(sourceType model.EvidenceSourceType) string {
	switch sourceType {
	case model.EvidenceSourceTypeURL:
		return evidence.SourceURL
	case model.EvidenceSourceTypeDocument:
		return evidence.SourceDocument
	case model.EvidenceSourceTypeNote:
		return evidence.SourceNote
	default:
		return evidence.SourceDataset
	}
}
