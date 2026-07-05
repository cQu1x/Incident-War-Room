package incident

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	// Create inserts a new incident and fills in the generated ID and CreatedAt.
	// Returns errs.ErrIncidentAlreadyActive if the chat already has an active incident.
	Create(ctx context.Context, inc *Incident) error

	// GetByID returns errs.ErrIncidentNotFound if no incident exists with the given ID.
	GetByID(ctx context.Context, id uuid.UUID) (*Incident, error)

	// List returns incidents ordered from newest to oldest. When chatID is
	// non-nil, only incidents belonging to that Telegram chat are returned.
	List(ctx context.Context, chatID *int64) ([]Incident, error)

	// GetActiveByTopicID returns the topic's active incident,
	// or errs.ErrNoActiveIncident if there is none.
	GetActiveByTopicID(ctx context.Context, chatID, topicID int64) (*Incident, error)

	// CountActiveByTitle returns how many active incidents in the chat carry
	// exactly the given title. It is used to keep active incident titles unique
	// within a chat by appending a numeric suffix on collision.
	CountActiveByTitle(ctx context.Context, chatID int64, title string) (int, error)

	// UpdateSeverity returns errs.ErrIncidentNotFound if no incident exists with the given ID.
	UpdateSeverity(ctx context.Context, id uuid.UUID, severity Severity) error

	UpdateTelegraphURLs(ctx context.Context, id uuid.UUID, telegraphURLs []string) error
	UpdateReportURL(ctx context.Context, id uuid.UUID, reportURL string) error

	// Close marks an active incident as closed.
	// Returns errs.ErrIncidentNotFound if the incident does not exist,
	// or errs.ErrIncidentAlreadyClosed if it is already closed.
	Close(ctx context.Context, id uuid.UUID, closedAt time.Time) error

	// Reopen brings a closed incident back to ACTIVE under newTopicID and clears
	// its closing time. Returns errs.ErrIncidentNotFound if the incident does not
	// exist, errs.ErrIncidentNotClosed if it is not closed, or
	// errs.ErrIncidentAlreadyActive if the chat already has an active incident in
	// that topic.
	Reopen(ctx context.Context, id uuid.UUID, newTopicID int64) error
}
