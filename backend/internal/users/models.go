// Package users owns user profiles: their identity, display name, and the
// password hash used by the auth domain. It is a leaf domain that other domains
// (teams, decisions) reference by user id.
package users

import (
	"time"

	"github.com/google/uuid"
)

// User is the canonical user entity. It carries no transport or storage tags;
// mapping to and from the database lives in the repository.
type User struct {
	ID           uuid.UUID
	Email        string
	Name         string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
