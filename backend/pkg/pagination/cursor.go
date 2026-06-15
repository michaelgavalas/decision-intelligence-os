// Package pagination provides Relay-style cursor pagination helpers. Cursors
// encode a (createdAt, id) position so that paging is stable under concurrent
// inserts.
package pagination

import (
	"encoding/base64"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/errors"
)

const (
	cursorSeparator = "|"
	defaultLimit    = 20
	maxLimit        = 100
)

// EncodeCursor encodes a (createdAt, id) position as a base64url string of the
// form "RFC3339Nano|uuid".
func EncodeCursor(createdAt time.Time, id uuid.UUID) string {
	raw := createdAt.UTC().Format(time.RFC3339Nano) + cursorSeparator + id.String()
	return base64.URLEncoding.EncodeToString([]byte(raw))
}

// DecodeCursor decodes a cursor produced by EncodeCursor back into its
// (createdAt, id) position.
func DecodeCursor(s string) (time.Time, uuid.UUID, error) {
	decoded, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return time.Time{}, uuid.Nil, errors.Wrap(err, errors.KindValidation, "INVALID_CURSOR", "cursor is not valid base64")
	}
	parts := strings.SplitN(string(decoded), cursorSeparator, 2)
	if len(parts) != 2 {
		return time.Time{}, uuid.Nil, errors.Validation("INVALID_CURSOR", "cursor is malformed")
	}
	createdAt, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return time.Time{}, uuid.Nil, errors.Wrap(err, errors.KindValidation, "INVALID_CURSOR", "cursor timestamp is invalid")
	}
	parsedID, err := uuid.Parse(parts[1])
	if err != nil {
		return time.Time{}, uuid.Nil, errors.Wrap(err, errors.KindValidation, "INVALID_CURSOR", "cursor id is invalid")
	}
	return createdAt, parsedID, nil
}

// PageArgs are the standard Relay forward/backward pagination arguments.
type PageArgs struct {
	First  *int
	After  *string
	Last   *int
	Before *string
}

// Validate enforces that at most one of First/Last is set and, when set, that
// it falls in the range 1..100. It returns a validation error otherwise.
func (p PageArgs) Validate() error {
	if p.First != nil && p.Last != nil {
		return errors.Validation("INVALID_PAGINATION", "first and last cannot be used together")
	}
	if p.First != nil && (*p.First < 1 || *p.First > maxLimit) {
		return errors.Validation("INVALID_PAGINATION", "first must be between 1 and 100")
	}
	if p.Last != nil && (*p.Last < 1 || *p.Last > maxLimit) {
		return errors.Validation("INVALID_PAGINATION", "last must be between 1 and 100")
	}
	return nil
}

// Limit returns the effective page size: First or Last when set, otherwise the
// default of 20, capped at 100.
func (p PageArgs) Limit() int {
	limit := defaultLimit
	switch {
	case p.First != nil:
		limit = *p.First
	case p.Last != nil:
		limit = *p.Last
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	if limit < 1 {
		limit = defaultLimit
	}
	return limit
}
