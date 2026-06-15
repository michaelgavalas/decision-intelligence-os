package ai

import (
	"context"
	"testing"

	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/errors"
)

func TestDisabled_AlwaysReportsFeatureDisabled(t *testing.T) {
	var provider LLMProvider = Disabled{}
	ctx := context.Background()

	if _, err := provider.Summarize(ctx, "text"); err != nil {
		assertCode(t, err, errors.KindValidation, "FEATURE_DISABLED")
	} else {
		t.Error("Summarize: expected FEATURE_DISABLED, got nil")
	}

	if _, err := provider.Critique(ctx, "text"); err != nil {
		assertCode(t, err, errors.KindValidation, "FEATURE_DISABLED")
	} else {
		t.Error("Critique: expected FEATURE_DISABLED, got nil")
	}

	report, err := provider.DetectBias(ctx, "text")
	if err != nil {
		assertCode(t, err, errors.KindValidation, "FEATURE_DISABLED")
	} else {
		t.Error("DetectBias: expected FEATURE_DISABLED, got nil")
	}
	if report.Summary != "" || len(report.Biases) != 0 {
		t.Errorf("DetectBias: expected zero-value report, got %+v", report)
	}
}
