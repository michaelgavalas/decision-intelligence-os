// Package clock abstracts the current time behind an interface so that code
// depending on time can be tested deterministically.
package clock

import "time"

// Clock reports the current time.
type Clock interface {
	Now() time.Time
}

// System is a Clock backed by the real wall clock.
type System struct{}

// Now returns the current wall-clock time.
func (System) Now() time.Time {
	return time.Now()
}

// Fixed is a Clock that always returns the same instant. It is intended for
// deterministic tests.
type Fixed struct {
	T time.Time
}

// Now returns the fixed instant.
func (f Fixed) Now() time.Time {
	return f.T
}
