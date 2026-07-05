package repository

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/cQu1x/Incident-War-Room/internal/domain/event"
	"github.com/cQu1x/Incident-War-Room/internal/errs"
)

// EventRepository stores timeline events in the "incident_events" table.
type EventRepository struct {
	db Querier
}

func NewEventRepository(db Querier) *EventRepository {
	return &EventRepository{db: db}
}

// Create inserts a new timeline event. ID and CreatedAt are generated
// by the database and written back into e.
func (r *EventRepository) Create(ctx context.Context, e *event.Event) error {
	const query = `
		INSERT INTO incident_events (incident_id, type, user_id, username, message, media_urls)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at`

	mediaURLs, err := marshalMediaURLs(e.MediaURLs)
	if err != nil {
		return errs.Wrapf(errs.KindInternal, "repository.Event.Create", err, "marshal media urls")
	}

	err = r.db.
		QueryRow(ctx, query, e.IncidentID, e.Type, e.UserID, e.Username, e.Message, mediaURLs).
		Scan(&e.ID, &e.CreatedAt)
	if err != nil {
		return errs.Wrapf(errs.KindInternal, "repository.Event.Create", err, "insert event")
	}

	return nil
}

// marshalMediaURLs encodes the media URLs as a JSON array for the media_urls
// TEXT column. An empty list is stored as SQL NULL to keep media-less events
// clean.
func marshalMediaURLs(urls []string) (any, error) {
	if len(urls) == 0 {
		return nil, nil
	}
	b, err := json.Marshal(urls)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

// ListByIncidentID returns all events of an incident in chronological
// order — this is the incident timeline.
func (r *EventRepository) ListByIncidentID(ctx context.Context, incidentID uuid.UUID) ([]event.Event, error) {
	const query = `
		SELECT id, incident_id, type, user_id, username, message, media_urls, created_at
		FROM incident_events
		WHERE incident_id = $1
		ORDER BY created_at ASC`

	rows, err := r.db.Query(ctx, query, incidentID)
	if err != nil {
		return nil, errs.Wrapf(errs.KindInternal, "repository.Event.ListByIncidentID", err, "select events")
	}

	events, err := pgx.CollectRows(rows, scanEvent)
	if err != nil {
		return nil, errs.Wrapf(errs.KindInternal, "repository.Event.ListByIncidentID", err, "scan events")
	}

	return events, nil
}

// ListParticipants returns distinct Telegram user IDs of everyone who
// produced at least one event in the incident. Events without a user
// (system events) are skipped.
func (r *EventRepository) ListParticipants(ctx context.Context, incidentID uuid.UUID) ([]int64, error) {
	const query = `
		SELECT DISTINCT user_id
		FROM incident_events
		WHERE incident_id = $1 AND user_id IS NOT NULL`

	rows, err := r.db.Query(ctx, query, incidentID)
	if err != nil {
		return nil, errs.Wrapf(errs.KindInternal, "repository.Event.ListParticipants", err, "select participants")
	}

	participants, err := pgx.CollectRows(rows, pgx.RowTo[int64])
	if err != nil {
		return nil, errs.Wrapf(errs.KindInternal, "repository.Event.ListParticipants", err, "scan participants")
	}

	return participants, nil
}

func scanEvent(row pgx.CollectableRow) (event.Event, error) {
	var e event.Event
	var mediaURLs *string
	err := row.Scan(
		&e.ID,
		&e.IncidentID,
		&e.Type,
		&e.UserID,
		&e.Username,
		&e.Message,
		&mediaURLs,
		&e.CreatedAt,
	)
	if err != nil {
		return e, err
	}
	if mediaURLs != nil && *mediaURLs != "" {
		if err := json.Unmarshal([]byte(*mediaURLs), &e.MediaURLs); err != nil {
			return e, err
		}
	}
	return e, nil
}
