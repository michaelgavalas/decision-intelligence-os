// Package id provides time-ordered UUIDv7 generation and parsing helpers used
// for primary keys across the backend.
package id

import "github.com/google/uuid"

// New returns a new time-ordered UUIDv7. It panics only on the practically
// impossible failure of the system random source.
func New() uuid.UUID {
	u, err := uuid.NewV7()
	if err != nil {
		panic("id: failed to generate UUIDv7: " + err.Error())
	}
	return u
}

// Parse parses a UUID from its string representation.
func Parse(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

// MustParse parses a UUID and panics if the input is invalid. It is intended
// for constants and tests where the input is known to be valid.
func MustParse(s string) uuid.UUID {
	return uuid.MustParse(s)
}
