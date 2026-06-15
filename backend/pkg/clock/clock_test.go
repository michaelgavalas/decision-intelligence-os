package clock_test

import (
	"testing"
	"time"

	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/clock"
)

func TestFixedReturnsT(t *testing.T) {
	want := time.Date(2026, 6, 14, 10, 0, 0, 0, time.UTC)
	var c clock.Clock = clock.Fixed{T: want}
	if got := c.Now(); !got.Equal(want) {
		t.Errorf("Fixed.Now() = %v, want %v", got, want)
	}
}

func TestSystemNow(t *testing.T) {
	var c clock.Clock = clock.System{}
	got := c.Now()
	delta := time.Since(got)
	if delta < 0 {
		delta = -delta
	}
	if delta > time.Second {
		t.Errorf("System.Now() off by %v, want within 1s of time.Now()", delta)
	}
}
