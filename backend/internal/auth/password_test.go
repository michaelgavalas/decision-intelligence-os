package auth

import (
	"strings"
	"testing"

	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/errors"
)

func TestHashAndVerifyPassword(t *testing.T) {
	encoded, err := HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if !strings.HasPrefix(encoded, "$argon2id$v=19$") {
		t.Errorf("encoded = %q, want PHC argon2id prefix", encoded)
	}

	ok, err := VerifyPassword(encoded, "correct horse battery staple")
	if err != nil {
		t.Fatalf("VerifyPassword: %v", err)
	}
	if !ok {
		t.Error("VerifyPassword = false, want true for the correct password")
	}
}

func TestVerifyPasswordWrongPassword(t *testing.T) {
	encoded, err := HashPassword("correct password")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}

	ok, err := VerifyPassword(encoded, "wrong password")
	if err != nil {
		t.Fatalf("VerifyPassword returned error on mismatch: %v", err)
	}
	if ok {
		t.Error("VerifyPassword = true, want false for the wrong password")
	}
}

func TestVerifyPasswordMalformed(t *testing.T) {
	cases := map[string]string{
		"garbage":          "not a phc string",
		"empty":            "",
		"wrong scheme":     "$argon2i$v=19$m=19456,t=2,p=1$c2FsdA$aGFzaA",
		"missing params":   "$argon2id$v=19$$c2FsdA$aGFzaA",
		"bad base64 salt":  "$argon2id$v=19$m=19456,t=2,p=1$!!!$aGFzaA",
		"bad base64 hash":  "$argon2id$v=19$m=19456,t=2,p=1$c2FsdA$!!!",
		"too few segments": "$argon2id$v=19$m=19456,t=2,p=1$c2FsdA",
	}
	for name, encoded := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := VerifyPassword(encoded, "whatever"); err == nil {
				t.Errorf("VerifyPassword(%q) err = nil, want error", encoded)
			}
		})
	}
}

func TestHashPasswordRandomSalt(t *testing.T) {
	a, err := HashPassword("same password")
	if err != nil {
		t.Fatalf("HashPassword a: %v", err)
	}
	b, err := HashPassword("same password")
	if err != nil {
		t.Fatalf("HashPassword b: %v", err)
	}
	if a == b {
		t.Error("two hashes of the same password are identical, want different salts")
	}
}

func TestHashPasswordEmptyRejected(t *testing.T) {
	_, err := HashPassword("")
	if errors.KindOf(err) != errors.KindValidation {
		t.Errorf("HashPassword(\"\") err = %v, want Validation", err)
	}
}
