package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/authctx"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/errors"
)

func TestAccessTokenRoundTrip(t *testing.T) {
	pub, priv, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}

	userID := uuid.New()
	teamID := uuid.New()
	now := time.Now()
	principal := authctx.Principal{UserID: userID, TeamID: &teamID, Role: "admin"}

	token, err := IssueAccessToken(priv, principal, now, time.Hour)
	if err != nil {
		t.Fatalf("IssueAccessToken: %v", err)
	}

	got, err := ParseAccessToken(pub, token)
	if err != nil {
		t.Fatalf("ParseAccessToken: %v", err)
	}
	if got.UserID != userID {
		t.Errorf("UserID = %v, want %v", got.UserID, userID)
	}
	if got.TeamID == nil || *got.TeamID != teamID {
		t.Errorf("TeamID = %v, want %v", got.TeamID, teamID)
	}
	if got.Role != "admin" {
		t.Errorf("Role = %q, want admin", got.Role)
	}
}

func TestParseAccessTokenExpired(t *testing.T) {
	pub, priv, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}

	past := time.Now().Add(-2 * time.Hour)
	token, err := IssueAccessToken(priv, authctx.Principal{UserID: uuid.New()}, past, time.Hour)
	if err != nil {
		t.Fatalf("IssueAccessToken: %v", err)
	}

	_, err = ParseAccessToken(pub, token)
	if errors.KindOf(err) != errors.KindUnauthenticated {
		t.Errorf("ParseAccessToken err = %v, want Unauthenticated", err)
	}
}

func TestParseAccessTokenWrongKey(t *testing.T) {
	_, priv, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}
	otherPub, _, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair other: %v", err)
	}

	token, err := IssueAccessToken(priv, authctx.Principal{UserID: uuid.New()}, time.Now(), time.Hour)
	if err != nil {
		t.Fatalf("IssueAccessToken: %v", err)
	}

	_, err = ParseAccessToken(otherPub, token)
	if errors.KindOf(err) != errors.KindUnauthenticated {
		t.Errorf("ParseAccessToken err = %v, want Unauthenticated", err)
	}
}

func TestParseAccessTokenTampered(t *testing.T) {
	pub, priv, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}

	token, err := IssueAccessToken(priv, authctx.Principal{UserID: uuid.New()}, time.Now(), time.Hour)
	if err != nil {
		t.Fatalf("IssueAccessToken: %v", err)
	}

	// Flip a character in the payload segment to invalidate the signature.
	tampered := token[:len(token)-3] + "AAA"
	_, err = ParseAccessToken(pub, tampered)
	if errors.KindOf(err) != errors.KindUnauthenticated {
		t.Errorf("ParseAccessToken err = %v, want Unauthenticated", err)
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	raw, hash, err := GenerateRefreshToken()
	if err != nil {
		t.Fatalf("GenerateRefreshToken: %v", err)
	}
	if raw == "" || hash == "" {
		t.Fatalf("GenerateRefreshToken returned empty raw=%q hash=%q", raw, hash)
	}
	if HashRefreshToken(raw) != hash {
		t.Errorf("HashRefreshToken(raw) = %q, want %q", HashRefreshToken(raw), hash)
	}

	raw2, _, err := GenerateRefreshToken()
	if err != nil {
		t.Fatalf("GenerateRefreshToken second: %v", err)
	}
	if raw == raw2 {
		t.Error("two refresh tokens are identical, want unique values")
	}
}
