// Package scalars implements custom GraphQL scalar marshaling for the API. The
// DateTime scalar maps Go's time.Time to and from an RFC 3339 string on the
// wire.
package scalars

import (
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/99designs/gqlgen/graphql"
)

// DateTime is the Go type bound to the GraphQL DateTime scalar. It is a plain
// time.Time; the marshaling functions below give gqlgen the wire format.
type DateTime = time.Time

// MarshalDateTime encodes a time.Time as an RFC 3339 (nanosecond precision)
// string for transport.
func MarshalDateTime(t time.Time) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		_, _ = io.WriteString(w, strconv.Quote(t.UTC().Format(time.RFC3339Nano)))
	})
}

// UnmarshalDateTime decodes an RFC 3339 string (or an already-decoded
// time.Time) into a time.Time.
func UnmarshalDateTime(v any) (time.Time, error) {
	switch val := v.(type) {
	case time.Time:
		return val, nil
	case string:
		t, err := time.Parse(time.RFC3339Nano, val)
		if err != nil {
			return time.Time{}, fmt.Errorf("DateTime must be a valid RFC 3339 timestamp: %w", err)
		}
		return t, nil
	default:
		return time.Time{}, fmt.Errorf("DateTime must be a string, got %T", v)
	}
}
