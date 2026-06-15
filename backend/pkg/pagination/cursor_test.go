package pagination_test

import (
	"testing"
	"time"

	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/errors"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/id"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/pagination"
)

func TestCursorRoundTrip(t *testing.T) {
	createdAt := time.Date(2026, 6, 14, 10, 30, 45, 123456789, time.UTC)
	u := id.New()

	encoded := pagination.EncodeCursor(createdAt, u)
	gotTime, gotID, err := pagination.DecodeCursor(encoded)
	if err != nil {
		t.Fatalf("DecodeCursor error: %v", err)
	}
	if !gotTime.Equal(createdAt) {
		t.Errorf("time = %v, want %v (sub-second precision lost)", gotTime, createdAt)
	}
	if gotID != u {
		t.Errorf("id = %v, want %v", gotID, u)
	}
}

func TestDecodeCursorMalformed(t *testing.T) {
	cases := []string{
		"",
		"not-base64-???",
		"bm90LWEtdmFsaWQtY3Vyc29y", // base64 of "not-a-valid-cursor" (no separator)
	}
	for _, c := range cases {
		t.Run(c, func(t *testing.T) {
			if _, _, err := pagination.DecodeCursor(c); err == nil {
				t.Errorf("DecodeCursor(%q) returned nil error, want error", c)
			}
		})
	}
}

func intPtr(i int) *int { return &i }

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		args    pagination.PageArgs
		wantErr bool
	}{
		{"empty ok", pagination.PageArgs{}, false},
		{"first ok", pagination.PageArgs{First: intPtr(10)}, false},
		{"last ok", pagination.PageArgs{Last: intPtr(10)}, false},
		{"first 1 ok", pagination.PageArgs{First: intPtr(1)}, false},
		{"first 100 ok", pagination.PageArgs{First: intPtr(100)}, false},
		{"first 0", pagination.PageArgs{First: intPtr(0)}, true},
		{"first 101", pagination.PageArgs{First: intPtr(101)}, true},
		{"last 0", pagination.PageArgs{Last: intPtr(0)}, true},
		{"last 101", pagination.PageArgs{Last: intPtr(101)}, true},
		{"both first and last", pagination.PageArgs{First: intPtr(5), Last: intPtr(5)}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.args.Validate()
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if errors.KindOf(err) != errors.KindValidation {
					t.Errorf("KindOf = %v, want KindValidation", errors.KindOf(err))
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestLimit(t *testing.T) {
	tests := []struct {
		name string
		args pagination.PageArgs
		want int
	}{
		{"default", pagination.PageArgs{}, 20},
		{"first", pagination.PageArgs{First: intPtr(15)}, 15},
		{"last", pagination.PageArgs{Last: intPtr(30)}, 30},
		{"caps first", pagination.PageArgs{First: intPtr(500)}, 100},
		{"caps last", pagination.PageArgs{Last: intPtr(500)}, 100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.Limit(); got != tt.want {
				t.Errorf("Limit() = %d, want %d", got, tt.want)
			}
		})
	}
}
