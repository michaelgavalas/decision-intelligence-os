package ai

import (
	"context"

	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/errors"
)

// Disabled is an LLMProvider that always reports the feature is unavailable. It
// is used as a safe default so the application can be constructed without any
// external AI configuration - for example when no API key is provided.
type Disabled struct{}

// disabledErr is the validation error returned by every Disabled method.
func disabledErr() error {
	return errors.Validation("FEATURE_DISABLED", "ai assistance is not enabled")
}

// Summarize always returns a FEATURE_DISABLED error.
func (Disabled) Summarize(_ context.Context, _ string) (string, error) {
	return "", disabledErr()
}

// Critique always returns a FEATURE_DISABLED error.
func (Disabled) Critique(_ context.Context, _ string) (string, error) {
	return "", disabledErr()
}

// DetectBias always returns a zero-value report and a FEATURE_DISABLED error.
func (Disabled) DetectBias(_ context.Context, _ string) (BiasReport, error) {
	return BiasReport{}, disabledErr()
}
