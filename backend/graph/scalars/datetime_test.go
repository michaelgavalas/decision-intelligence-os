package scalars_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/michaelgavalas/decision-intelligence-os/backend/graph/scalars"
)

func TestMarshalDateTime(t *testing.T) {
	ts := time.Date(2026, 6, 14, 9, 30, 0, 0, time.UTC)

	var buf bytes.Buffer
	scalars.MarshalDateTime(ts).MarshalGQL(&buf)

	want := `"2026-06-14T09:30:00Z"`
	if got := buf.String(); got != want {
		t.Errorf("MarshalDateTime = %s, want %s", got, want)
	}
}

func TestMarshalDateTimeConvertsToUTC(t *testing.T) {
	loc := time.FixedZone("UTC+2", 2*60*60)
	ts := time.Date(2026, 6, 14, 11, 30, 0, 0, loc)

	var buf bytes.Buffer
	scalars.MarshalDateTime(ts).MarshalGQL(&buf)

	want := `"2026-06-14T09:30:00Z"`
	if got := buf.String(); got != want {
		t.Errorf("MarshalDateTime = %s, want %s", got, want)
	}
}

func TestUnmarshalDateTimeFromString(t *testing.T) {
	got, err := scalars.UnmarshalDateTime("2026-06-14T09:30:00Z")
	if err != nil {
		t.Fatalf("UnmarshalDateTime returned error: %v", err)
	}
	want := time.Date(2026, 6, 14, 9, 30, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("UnmarshalDateTime = %v, want %v", got, want)
	}
}

func TestUnmarshalDateTimeFromTime(t *testing.T) {
	want := time.Date(2026, 6, 14, 9, 30, 0, 0, time.UTC)
	got, err := scalars.UnmarshalDateTime(want)
	if err != nil {
		t.Fatalf("UnmarshalDateTime returned error: %v", err)
	}
	if !got.Equal(want) {
		t.Errorf("UnmarshalDateTime = %v, want %v", got, want)
	}
}

func TestUnmarshalDateTimeInvalidString(t *testing.T) {
	if _, err := scalars.UnmarshalDateTime("not a timestamp"); err == nil {
		t.Error("UnmarshalDateTime(invalid) returned nil error, want error")
	}
}

func TestUnmarshalDateTimeWrongType(t *testing.T) {
	if _, err := scalars.UnmarshalDateTime(42); err == nil {
		t.Error("UnmarshalDateTime(int) returned nil error, want error")
	}
}
