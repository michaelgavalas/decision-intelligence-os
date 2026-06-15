// Package evidence owns the evidence attached to an assumption: the sources and
// notes that support or challenge a belief. Evidence is scoped to an assumption
// and authorized through it.
package evidence

import (
	"time"

	"github.com/google/uuid"
)

// Evidence source types.
const (
	// SourceURL is a link to an external resource.
	SourceURL = "url"
	// SourceDocument is an uploaded or referenced document.
	SourceDocument = "document"
	// SourceNote is a free-form written note.
	SourceNote = "note"
	// SourceDataset is a structured dataset.
	SourceDataset = "dataset"
)

// ValidSourceType reports whether s is a recognized evidence source type.
func ValidSourceType(s string) bool {
	switch s {
	case SourceURL, SourceDocument, SourceNote, SourceDataset:
		return true
	default:
		return false
	}
}

// Evidence is the canonical evidence entity. SourceURL is optional and present
// chiefly for url-typed evidence.
type Evidence struct {
	ID           uuid.UUID
	AssumptionID uuid.UUID
	SourceType   string
	SourceURL    *string
	Content      string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
