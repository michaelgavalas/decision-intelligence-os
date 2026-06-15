package main

import (
	"crypto/ed25519"
	"encoding/base64"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/auth"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/config"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/authctx"
)

func TestLoadTokenKeys_DevEmptyGeneratesWorkingPair(t *testing.T) {
	cfg := config.Config{Env: "development"}

	pub, priv, err := loadTokenKeys(cfg, testLogger())
	if err != nil {
		t.Fatalf("loadTokenKeys: %v", err)
	}
	if len(pub) != ed25519.PublicKeySize || len(priv) != ed25519.PrivateKeySize {
		t.Fatalf("unexpected key sizes: pub=%d priv=%d", len(pub), len(priv))
	}

	assertSignVerify(t, pub, priv)
}

func TestLoadTokenKeys_ProductionEmptyIsFatal(t *testing.T) {
	cfg := config.Config{Env: "production"}

	if _, _, err := loadTokenKeys(cfg, testLogger()); err == nil {
		t.Fatal("expected error for missing keys in production")
	}
}

func TestLoadTokenKeys_Base64RoundTrips(t *testing.T) {
	_, priv, err := auth.GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}

	cfg := config.Config{
		Env:           "production",
		JWTPrivateKey: base64.StdEncoding.EncodeToString(priv),
	}

	pub, gotPriv, err := loadTokenKeys(cfg, testLogger())
	if err != nil {
		t.Fatalf("loadTokenKeys: %v", err)
	}
	if !priv.Equal(gotPriv) {
		t.Fatal("decoded private key does not match input")
	}

	assertSignVerify(t, pub, gotPriv)
}

func TestLoadTokenKeys_PublicKeyMismatch(t *testing.T) {
	_, priv, _ := auth.GenerateKeyPair()
	otherPub, _, _ := auth.GenerateKeyPair()

	cfg := config.Config{
		Env:           "production",
		JWTPrivateKey: base64.StdEncoding.EncodeToString(priv),
		JWTPublicKey:  base64.StdEncoding.EncodeToString(otherPub),
	}

	if _, _, err := loadTokenKeys(cfg, testLogger()); err == nil {
		t.Fatal("expected mismatch error")
	}
}

// assertSignVerify confirms the keypair can issue and parse an access token.
func assertSignVerify(t *testing.T, pub ed25519.PublicKey, priv ed25519.PrivateKey) {
	t.Helper()

	want := authctx.Principal{UserID: uuid.New(), Role: "admin"}
	token, err := auth.IssueAccessToken(priv, want, time.Now(), time.Minute)
	if err != nil {
		t.Fatalf("IssueAccessToken: %v", err)
	}

	got, err := auth.ParseAccessToken(pub, token)
	if err != nil {
		t.Fatalf("ParseAccessToken: %v", err)
	}
	if got.UserID != want.UserID || got.Role != want.Role {
		t.Fatalf("principal mismatch: got %+v want %+v", got, want)
	}
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
