package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/cQu1x/Incident-War-Room/internal/domain/event"
	"github.com/cQu1x/Incident-War-Room/internal/domain/incident"
)

// ReopenIncident reopens a previously closed incident under newTopicID, keeping
// its existing timeline intact, and records an INCIDENT_REOPENED event on it.
// Flipping the incident back to ACTIVE and writing the event happen in one
// transaction.
//
// Returns errs.ErrIncidentNotFound if no incident has the given ID,
// errs.ErrIncidentNotClosed if it is not closed, or
// errs.ErrIncidentAlreadyActive if the chat already has an active incident in
// that topic.
func (s *Service) ReopenIncident(
	ctx context.Context,
	id uuid.UUID,
	newTopicID int64,
	userID *int64,
	username string,
) (*incident.Incident, error) {
	var reopened *incident.Incident
	err := s.tx.WithTx(ctx, func(incidents incident.Repository, events event.Repository) error {
		if err := incidents.Reopen(ctx, id, newTopicID); err != nil {
			return err
		}

		if err := events.Create(ctx, &event.Event{
			IncidentID: id,
			Type:       event.TypeIncidentReopened,
			UserID:     userID,
			Username:   username,
		}); err != nil {
			return err
		}

		inc, err := incidents.GetByID(ctx, id)
		if err != nil {
			return err
		}
		reopened = inc
		return nil
	})
	if err != nil {
		return nil, err
	}

	return reopened, nil
}
