package ai

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/authctx"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/errors"
)

// fakeProvider records the calls made to it and returns canned values. Each
// method fails the test if it is invoked when the service is supposed to
// short-circuit (e.g. when AI assistance is disabled).
type fakeProvider struct {
	t *testing.T

	summarizeCalls int
	critiqueCalls  int
	detectCalls    int

	failIfCalled bool

	summary  string
	critique string
	report   BiasReport
}

func (f *fakeProvider) Summarize(_ context.Context, _ string) (string, error) {
	if f.failIfCalled {
		f.t.Fatal("provider.Summarize called when it should not have been")
	}
	f.summarizeCalls++
	return f.summary, nil
}

func (f *fakeProvider) Critique(_ context.Context, _ string) (string, error) {
	if f.failIfCalled {
		f.t.Fatal("provider.Critique called when it should not have been")
	}
	f.critiqueCalls++
	return f.critique, nil
}

func (f *fakeProvider) DetectBias(_ context.Context, _ string) (BiasReport, error) {
	if f.failIfCalled {
		f.t.Fatal("provider.DetectBias called when it should not have been")
	}
	f.detectCalls++
	return f.report, nil
}

// ctxWithPrincipal returns a context carrying an authenticated principal.
func ctxWithPrincipal() context.Context {
	return authctx.WithPrincipal(context.Background(), authctx.Principal{
		UserID: uuid.New(),
		Role:   "MEMBER",
	})
}

func assertCode(t *testing.T, err error, wantKind errors.Kind, wantCode string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error with code %q, got nil", wantCode)
	}
	if got := errors.KindOf(err); got != wantKind {
		t.Errorf("kind = %v, want %v", got, wantKind)
	}
	if got := errors.CodeOf(err); got != wantCode {
		t.Errorf("code = %q, want %q", got, wantCode)
	}
}

func TestService_Disabled_ReturnsFeatureDisabledAndNeverCallsProvider(t *testing.T) {
	provider := &fakeProvider{t: t, failIfCalled: true}
	svc := NewService(false, provider)
	ctx := ctxWithPrincipal()

	if svc.Enabled() {
		t.Error("Enabled() = true, want false")
	}

	if _, err := svc.SummarizeEvidence(ctx, "some evidence"); err != nil {
		assertCode(t, err, errors.KindValidation, "FEATURE_DISABLED")
	} else {
		t.Error("SummarizeEvidence: expected FEATURE_DISABLED, got nil")
	}

	if _, err := svc.CritiqueAssumption(ctx, "some assumption"); err != nil {
		assertCode(t, err, errors.KindValidation, "FEATURE_DISABLED")
	} else {
		t.Error("CritiqueAssumption: expected FEATURE_DISABLED, got nil")
	}

	report, err := svc.DetectBias(ctx, "some text")
	if err != nil {
		assertCode(t, err, errors.KindValidation, "FEATURE_DISABLED")
	} else {
		t.Error("DetectBias: expected FEATURE_DISABLED, got nil")
	}
	if report.Summary != "" || len(report.Biases) != 0 {
		t.Errorf("DetectBias: expected zero-value report, got %+v", report)
	}
}

func TestService_MissingPrincipal_ReturnsUnauthenticated(t *testing.T) {
	provider := &fakeProvider{t: t, failIfCalled: true}
	svc := NewService(true, provider)
	ctx := context.Background()

	if _, err := svc.SummarizeEvidence(ctx, "x"); err != nil {
		assertCode(t, err, errors.KindUnauthenticated, "UNAUTHENTICATED")
	} else {
		t.Error("SummarizeEvidence: expected UNAUTHENTICATED, got nil")
	}

	if _, err := svc.CritiqueAssumption(ctx, "x"); err != nil {
		assertCode(t, err, errors.KindUnauthenticated, "UNAUTHENTICATED")
	} else {
		t.Error("CritiqueAssumption: expected UNAUTHENTICATED, got nil")
	}

	if _, err := svc.DetectBias(ctx, "x"); err != nil {
		assertCode(t, err, errors.KindUnauthenticated, "UNAUTHENTICATED")
	} else {
		t.Error("DetectBias: expected UNAUTHENTICATED, got nil")
	}
}

func TestService_EmptyInput_ReturnsEmptyInput(t *testing.T) {
	provider := &fakeProvider{t: t, failIfCalled: true}
	svc := NewService(true, provider)
	ctx := ctxWithPrincipal()

	if _, err := svc.SummarizeEvidence(ctx, "   "); err != nil {
		assertCode(t, err, errors.KindValidation, "EMPTY_INPUT")
	} else {
		t.Error("SummarizeEvidence: expected EMPTY_INPUT, got nil")
	}

	if _, err := svc.CritiqueAssumption(ctx, ""); err != nil {
		assertCode(t, err, errors.KindValidation, "EMPTY_INPUT")
	} else {
		t.Error("CritiqueAssumption: expected EMPTY_INPUT, got nil")
	}

	if _, err := svc.DetectBias(ctx, "\t\n"); err != nil {
		assertCode(t, err, errors.KindValidation, "EMPTY_INPUT")
	} else {
		t.Error("DetectBias: expected EMPTY_INPUT, got nil")
	}
}

func TestService_Enabled_DelegatesToProvider(t *testing.T) {
	provider := &fakeProvider{
		t:        t,
		summary:  "a concise summary",
		critique: "a thorough critique",
		report: BiasReport{
			Summary: "one bias found",
			Biases:  []DetectedBias{{Name: "Anchoring", Explanation: "fixated on the first number"}},
		},
	}
	svc := NewService(true, provider)
	ctx := ctxWithPrincipal()

	if !svc.Enabled() {
		t.Error("Enabled() = false, want true")
	}

	got, err := svc.SummarizeEvidence(ctx, "evidence text")
	if err != nil {
		t.Fatalf("SummarizeEvidence: unexpected error: %v", err)
	}
	if got != provider.summary {
		t.Errorf("SummarizeEvidence = %q, want %q", got, provider.summary)
	}
	if provider.summarizeCalls != 1 {
		t.Errorf("summarizeCalls = %d, want 1", provider.summarizeCalls)
	}

	gotCritique, err := svc.CritiqueAssumption(ctx, "assumption text")
	if err != nil {
		t.Fatalf("CritiqueAssumption: unexpected error: %v", err)
	}
	if gotCritique != provider.critique {
		t.Errorf("CritiqueAssumption = %q, want %q", gotCritique, provider.critique)
	}
	if provider.critiqueCalls != 1 {
		t.Errorf("critiqueCalls = %d, want 1", provider.critiqueCalls)
	}

	gotReport, err := svc.DetectBias(ctx, "biased text")
	if err != nil {
		t.Fatalf("DetectBias: unexpected error: %v", err)
	}
	if gotReport.Summary != provider.report.Summary {
		t.Errorf("DetectBias summary = %q, want %q", gotReport.Summary, provider.report.Summary)
	}
	if len(gotReport.Biases) != 1 || gotReport.Biases[0].Name != "Anchoring" {
		t.Errorf("DetectBias biases = %+v, want one Anchoring bias", gotReport.Biases)
	}
	if provider.detectCalls != 1 {
		t.Errorf("detectCalls = %d, want 1", provider.detectCalls)
	}
}
