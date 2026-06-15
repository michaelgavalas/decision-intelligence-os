package ai

import (
	"context"
	"strings"

	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/authctx"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/errors"
)

// Service exposes the optional AI-assistance features to the rest of the
// application. When AI is disabled (the default), every method returns a
// FEATURE_DISABLED error and the provider is never contacted, so core workflows
// remain fully functional.
type Service struct {
	enabled  bool
	provider LLMProvider
}

// NewService wires a Service from an enabled flag and a provider. The provider
// is only ever contacted when enabled is true; callers may pass a Disabled
// provider when the feature is off.
func NewService(enabled bool, provider LLMProvider) *Service {
	return &Service{enabled: enabled, provider: provider}
}

// Enabled reports whether the AI-assistance feature is turned on.
func (s *Service) Enabled() bool {
	return s.enabled
}

// SummarizeEvidence returns a concise summary of a piece of evidence text. It
// requires an authenticated caller and the feature to be enabled.
func (s *Service) SummarizeEvidence(ctx context.Context, text string) (string, error) {
	if err := s.guard(ctx, text); err != nil {
		return "", err
	}
	return s.provider.Summarize(ctx, text)
}

// CritiqueAssumption returns a critical evaluation of an assumption statement. It
// requires an authenticated caller and the feature to be enabled.
func (s *Service) CritiqueAssumption(ctx context.Context, statement string) (string, error) {
	if err := s.guard(ctx, statement); err != nil {
		return "", err
	}
	return s.provider.Critique(ctx, statement)
}

// DetectBias returns the cognitive biases the model identified in the text. It
// requires an authenticated caller and the feature to be enabled; otherwise it
// returns a zero-value report alongside the error.
func (s *Service) DetectBias(ctx context.Context, text string) (BiasReport, error) {
	if err := s.guard(ctx, text); err != nil {
		return BiasReport{}, err
	}
	return s.provider.DetectBias(ctx, text)
}

// guard enforces the common preconditions for every AI-assistance call:
// authentication, the feature being enabled, and non-empty input.
func (s *Service) guard(ctx context.Context, input string) error {
	if _, err := authctx.Require(ctx); err != nil {
		return err
	}
	if !s.enabled {
		return errors.Validation("FEATURE_DISABLED", "ai assistance is not enabled")
	}
	if strings.TrimSpace(input) == "" {
		return errors.Validation("EMPTY_INPUT", "input must not be empty")
	}
	return nil
}
