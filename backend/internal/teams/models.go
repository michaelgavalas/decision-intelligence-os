// Package teams owns team identity and membership: which users belong to a team
// and in what role. Membership roles gate authorization across the product.
package teams

import (
	"time"

	"github.com/google/uuid"
)

// Membership roles, ordered from most to least privileged.
const (
	// RoleAdmin can manage membership and team settings.
	RoleAdmin = "admin"
	// RoleMember can contribute within the team.
	RoleMember = "member"
	// RoleViewer has read-only access.
	RoleViewer = "viewer"
)

// ValidRole reports whether r is a recognized membership role.
func ValidRole(r string) bool {
	switch r {
	case RoleAdmin, RoleMember, RoleViewer:
		return true
	default:
		return false
	}
}

// Team is the canonical team entity.
type Team struct {
	ID        uuid.UUID
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Membership links a user to a team with a role.
type Membership struct {
	TeamID    uuid.UUID
	UserID    uuid.UUID
	Role      string
	CreatedAt time.Time
}
