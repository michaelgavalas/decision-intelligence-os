package id_test

import (
	"testing"

	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/id"
)

func TestNewIsVersion7(t *testing.T) {
	u := id.New()
	if u.Version() != 7 {
		t.Errorf("Version() = %d, want 7", u.Version())
	}
}

func TestNewIsTimeOrdered(t *testing.T) {
	const n = 100
	prev := id.New().String()
	for i := 1; i < n; i++ {
		cur := id.New().String()
		if cur < prev {
			t.Fatalf("ids not non-decreasing at %d: %q < %q", i, cur, prev)
		}
		prev = cur
	}
}

func TestParseRoundTrip(t *testing.T) {
	u := id.New()
	parsed, err := id.Parse(u.String())
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if parsed != u {
		t.Errorf("Parse round-trip = %v, want %v", parsed, u)
	}
}

func TestParseGarbage(t *testing.T) {
	if _, err := id.Parse("not-a-uuid"); err == nil {
		t.Error("Parse(garbage) returned nil error, want error")
	}
}

func TestMustParse(t *testing.T) {
	u := id.New()
	if got := id.MustParse(u.String()); got != u {
		t.Errorf("MustParse = %v, want %v", got, u)
	}

	defer func() {
		if recover() == nil {
			t.Error("MustParse(garbage) did not panic")
		}
	}()
	id.MustParse("not-a-uuid")
}
