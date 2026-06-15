// Package ai implements the optional AI-assistance product feature. It lets end
// users request AI help on their decision work - summarizing evidence,
// critiquing an assumption, and detecting cognitive bias - by integrating an
// external large-language-model service.
//
// The feature is disabled by default. When it is off, the Service short-circuits
// before contacting any provider, so the core decision-quality workflows remain
// fully functional without any AI configuration.
package ai

import "context"

// DetectedBias is one cognitive bias the model flagged in a piece of text.
type DetectedBias struct {
	Name        string
	Explanation string
}

// BiasReport is the result of a bias-detection request: a short summary plus the
// individual biases the model identified.
type BiasReport struct {
	Summary string
	Biases  []DetectedBias
}

// LLMProvider is the integration boundary for the optional AI-assistance
// features. Implementations call an external large-language-model service, the
// same way an application integrates a payment or email vendor. Swapping the
// provider (real, disabled, or a test double) requires no changes to callers.
type LLMProvider interface {
	// Summarize returns a concise, neutral summary of the supplied text.
	Summarize(ctx context.Context, text string) (string, error)
	// Critique returns a critical evaluation of the supplied statement.
	Critique(ctx context.Context, statement string) (string, error)
	// DetectBias returns the cognitive biases the model identified in the text.
	DetectBias(ctx context.Context, text string) (BiasReport, error)
}
