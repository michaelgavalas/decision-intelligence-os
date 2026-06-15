package main

import (
	"crypto/ed25519"
	"encoding/base64"
	"log/slog"

	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/auth"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/config"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/errors"
)

// loadTokenKeys returns the signing keypair for access tokens. When both keys
// are present in cfg they are decoded from standard base64-encoded raw key
// bytes. In development, when keys are absent, an ephemeral keypair is generated
// (with a warning) so the application runs locally without configuration. In
// production, absent keys are a fatal configuration error.
func loadTokenKeys(cfg config.Config, log *slog.Logger) (ed25519.PublicKey, ed25519.PrivateKey, error) {
	if cfg.JWTPrivateKey == "" {
		if cfg.IsProduction() {
			return nil, nil, errors.Validation(
				"TOKEN_KEYS_MISSING",
				"JWT signing keys are required in production",
			)
		}
		log.Warn("token signing keys not configured; generating an ephemeral keypair (development only)")
		return auth.GenerateKeyPair()
	}

	privBytes, err := base64.StdEncoding.DecodeString(cfg.JWTPrivateKey)
	if err != nil {
		return nil, nil, errors.Wrap(err, errors.KindValidation, "TOKEN_PRIVATE_KEY_INVALID", "JWT private key is not valid base64")
	}
	if len(privBytes) != ed25519.PrivateKeySize {
		return nil, nil, errors.Validation("TOKEN_PRIVATE_KEY_INVALID", "JWT private key must be 64 raw bytes")
	}
	priv := ed25519.PrivateKey(privBytes)

	// Derive the public key from the private key by default; only decode the
	// configured public key when one is provided, and verify it matches.
	pub, ok := priv.Public().(ed25519.PublicKey)
	if !ok {
		return nil, nil, errors.Internal("TOKEN_PRIVATE_KEY_INVALID", "JWT private key did not yield an ed25519 public key")
	}

	if cfg.JWTPublicKey != "" {
		pubBytes, err := base64.StdEncoding.DecodeString(cfg.JWTPublicKey)
		if err != nil {
			return nil, nil, errors.Wrap(err, errors.KindValidation, "TOKEN_PUBLIC_KEY_INVALID", "JWT public key is not valid base64")
		}
		if len(pubBytes) != ed25519.PublicKeySize {
			return nil, nil, errors.Validation("TOKEN_PUBLIC_KEY_INVALID", "JWT public key must be 32 raw bytes")
		}
		if !ed25519.PublicKey(pubBytes).Equal(pub) {
			return nil, nil, errors.Validation("TOKEN_KEYS_MISMATCH", "configured JWT public key does not match the private key")
		}
	}

	return pub, priv, nil
}
