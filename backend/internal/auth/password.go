// Package auth owns authentication: password hashing, access and refresh token
// issuance, and the registration, login, refresh, and logout flows. It depends
// on the users and teams domains through narrow interfaces so it stays decoupled
// from their internals.
package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	stderrors "errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"

	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/errors"
)

// argon2id parameters. These are moderate, interactive-login costs: roughly
// 19 MiB of memory and two passes, which resist GPU cracking while keeping
// login latency acceptable. They are encoded into every hash so they can be
// tuned later without breaking existing passwords.
const (
	argonMemoryKiB = 19456
	argonTime      = 2
	argonThreads   = 1
	argonKeyLen    = 32
	argonSaltLen   = 16
)

// errMalformedHash is returned when an encoded password cannot be parsed.
var errMalformedHash = stderrors.New("auth: malformed password hash")

// HashPassword derives an argon2id hash of plain and returns it as a PHC-format
// string ($argon2id$v=19$m=...,t=...,p=...$salt$hash) with a fresh random salt.
// An empty password is rejected as a validation error.
func HashPassword(plain string) (string, error) {
	if plain == "" {
		return "", errors.Validation("WEAK_PASSWORD", "password must not be empty")
	}

	salt := make([]byte, argonSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", errors.Wrap(err, errors.KindInternal, "PASSWORD_HASH_FAILED", "failed to read random salt")
	}

	hash := argon2.IDKey([]byte(plain), salt, argonTime, argonMemoryKiB, argonThreads, argonKeyLen)

	encoded := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		argonMemoryKiB, argonTime, argonThreads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)
	return encoded, nil
}

// VerifyPassword reports whether plain matches the PHC-encoded argon2id hash. A
// non-matching password returns (false, nil); only a malformed encoded value
// returns an error. The comparison is constant time.
func VerifyPassword(encoded, plain string) (bool, error) {
	memory, time, threads, salt, want, err := decodeHash(encoded)
	if err != nil {
		return false, err
	}

	got := argon2.IDKey([]byte(plain), salt, time, memory, threads, uint32(len(want)))
	return subtle.ConstantTimeCompare(got, want) == 1, nil
}

// decodeHash parses a PHC argon2id string into its parameters, salt, and hash.
func decodeHash(encoded string) (memory, time uint32, threads uint8, salt, hash []byte, err error) {
	parts := strings.Split(encoded, "$")
	// A well-formed value splits into ["", "argon2id", "v=19", "m=..,t=..,p=..",
	// "<salt>", "<hash>"].
	if len(parts) != 6 || parts[0] != "" || parts[1] != "argon2id" {
		return 0, 0, 0, nil, nil, errMalformedHash
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil || version != argon2.Version {
		return 0, 0, 0, nil, nil, errMalformedHash
	}

	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &threads); err != nil {
		return 0, 0, 0, nil, nil, errMalformedHash
	}

	salt, err = base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil || len(salt) == 0 {
		return 0, 0, 0, nil, nil, errMalformedHash
	}

	hash, err = base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil || len(hash) == 0 {
		return 0, 0, 0, nil, nil, errMalformedHash
	}

	return memory, time, threads, salt, hash, nil
}
