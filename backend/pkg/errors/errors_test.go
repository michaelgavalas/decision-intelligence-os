package errors_test

import (
	stderrors "errors"
	"fmt"
	"strings"
	"testing"

	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/errors"
)

func TestConstructors(t *testing.T) {
	tests := []struct {
		name string
		err  *errors.Error
		kind errors.Kind
	}{
		{"not found", errors.NotFound("USER_NOT_FOUND", "user not found"), errors.KindNotFound},
		{"validation", errors.Validation("EMAIL_INVALID", "email invalid"), errors.KindValidation},
		{"conflict", errors.Conflict("EMAIL_TAKEN", "email taken"), errors.KindConflict},
		{"unauthenticated", errors.Unauthenticated("NO_SESSION", "no session"), errors.KindUnauthenticated},
		{"forbidden", errors.Forbidden("NO_ACCESS", "no access"), errors.KindForbidden},
		{"internal", errors.Internal("BOOM", "boom"), errors.KindInternal},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Kind != tt.kind {
				t.Errorf("Kind = %v, want %v", tt.err.Kind, tt.kind)
			}
			if tt.err.Code == "" {
				t.Error("Code is empty")
			}
			if tt.err.Message == "" {
				t.Error("Message is empty")
			}
			if !strings.Contains(tt.err.Error(), tt.err.Code) {
				t.Errorf("Error() = %q, want it to contain Code %q", tt.err.Error(), tt.err.Code)
			}
			if !strings.Contains(tt.err.Error(), tt.err.Message) {
				t.Errorf("Error() = %q, want it to contain Message %q", tt.err.Error(), tt.err.Message)
			}
		})
	}
}

func TestKindOfPlainError(t *testing.T) {
	err := fmt.Errorf("some plain error")
	if got := errors.KindOf(err); got != errors.KindInternal {
		t.Errorf("KindOf(plain) = %v, want KindInternal", got)
	}
	if got := errors.CodeOf(err); got != "" {
		t.Errorf("CodeOf(plain) = %q, want empty", got)
	}
	if got := errors.KindOf(nil); got != errors.KindInternal {
		t.Errorf("KindOf(nil) = %v, want KindInternal", got)
	}
}

func TestWrapPreservesUnwrap(t *testing.T) {
	sentinel := stderrors.New("db connection refused")
	wrapped := errors.Wrap(sentinel, errors.KindInternal, "DB_DOWN", "database unavailable")

	if !stderrors.Is(wrapped, sentinel) {
		t.Error("errors.Is could not find the wrapped sentinel")
	}
	if !strings.Contains(wrapped.Error(), sentinel.Error()) {
		t.Errorf("Error() = %q, want it to append wrapped error %q", wrapped.Error(), sentinel.Error())
	}
	if wrapped.Unwrap() != sentinel {
		t.Error("Unwrap did not return the sentinel")
	}
}

func TestKindAndCodeThroughChain(t *testing.T) {
	base := errors.Conflict("EMAIL_TAKEN", "email already in use")
	chain := fmt.Errorf("service layer: %w", base)
	chain = fmt.Errorf("handler: %w", chain)

	if got := errors.KindOf(chain); got != errors.KindConflict {
		t.Errorf("KindOf(chain) = %v, want KindConflict", got)
	}
	if got := errors.CodeOf(chain); got != "EMAIL_TAKEN" {
		t.Errorf("CodeOf(chain) = %q, want EMAIL_TAKEN", got)
	}
}
