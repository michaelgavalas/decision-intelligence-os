package graph_test

import (
	"context"
	stderrors "errors"
	"fmt"
	"testing"

	"github.com/michaelgavalas/decision-intelligence-os/backend/graph"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/errors"
)

func TestPresentErrorKindMapping(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode string
		wantMsg  string
	}{
		{
			name:     "unauthenticated",
			err:      errors.Unauthenticated("NO_SESSION", "authentication required"),
			wantCode: "UNAUTHENTICATED",
			wantMsg:  "authentication required",
		},
		{
			name:     "forbidden",
			err:      errors.Forbidden("NO_ACCESS", "access denied"),
			wantCode: "FORBIDDEN",
			wantMsg:  "access denied",
		},
		{
			name:     "not found",
			err:      errors.NotFound("USER_NOT_FOUND", "user not found"),
			wantCode: "NOT_FOUND",
			wantMsg:  "user not found",
		},
		{
			name:     "conflict",
			err:      errors.Conflict("EMAIL_TAKEN", "email already in use"),
			wantCode: "CONFLICT",
			wantMsg:  "email already in use",
		},
		{
			name:     "validation",
			err:      errors.Validation("EMAIL_INVALID", "email is invalid"),
			wantCode: "VALIDATION",
			wantMsg:  "email is invalid",
		},
		{
			name:     "internal hides cause",
			err:      errors.Internal("BOOM", "the database exploded"),
			wantCode: "INTERNAL",
			wantMsg:  "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := graph.PresentError(context.Background(), tt.err)

			if got.Message != tt.wantMsg {
				t.Errorf("Message = %q, want %q", got.Message, tt.wantMsg)
			}
			if code, _ := got.Extensions["code"].(string); code != tt.wantCode {
				t.Errorf("extensions[code] = %v, want %q", got.Extensions["code"], tt.wantCode)
			}
			reason, _ := got.Extensions["reason"].(string)
			if reason != errors.CodeOf(tt.err) {
				t.Errorf("extensions[reason] = %q, want %q", reason, errors.CodeOf(tt.err))
			}
		})
	}
}

func TestPresentErrorInternalDoesNotLeakCause(t *testing.T) {
	cause := stderrors.New("connection refused at 10.0.0.5:5432")
	wrapped := errors.Wrap(cause, errors.KindInternal, "DB_DOWN", "database unavailable")

	got := graph.PresentError(context.Background(), wrapped)

	if got.Message != "internal server error" {
		t.Errorf("Message = %q, want %q", got.Message, "internal server error")
	}
	if got.Message == wrapped.Error() {
		t.Error("internal error leaked its full error string")
	}
}

func TestPresentErrorPlainErrorIsInternal(t *testing.T) {
	got := graph.PresentError(context.Background(), fmt.Errorf("some unclassified failure"))

	if code, _ := got.Extensions["code"].(string); code != "INTERNAL" {
		t.Errorf("extensions[code] = %v, want INTERNAL", got.Extensions["code"])
	}
	if got.Message != "internal server error" {
		t.Errorf("Message = %q, want %q", got.Message, "internal server error")
	}
	if _, ok := got.Extensions["reason"]; ok {
		t.Error("plain error should not carry a reason extension")
	}
}
