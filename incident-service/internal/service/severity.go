package service

import (
	"context"
	"fmt"

	"github.com/cQu1x/Incident-War-Room/internal/domain/event"
	"github.com/cQu1x/Incident-War-Room/internal/domain/incident"
	"github.com/cQu1x/Incident-War-Room/internal/errs"
)

// SetSeverity changes the severity of the chat's active incident and records a
// SEVERITY_CHANGED event describing the transition on its timeline. Updating the
// severity and writing the event happen in one transaction. Setting the same
// severity is a no-op and records no event.
//
// Returns errs.ErrNoActiveIncident if the chat has no active incident, or an
// errs.KindValidation error if the severity is invalid.
func (s *Service) SetSeverity(ctx context.Context, chatID, topicID int64, severity incident.Severity, userID *int64, username string) (*incident.Incident, error) {
	const op = "service.SetSeverity"

	if !validSeverity(severity) {
		return nil, errs.New(errs.KindValidation, op, "invalid severity")
	}

	var updated *incident.Incident
	err := s.tx.WithTx(ctx, func(incidents incident.Repository, events event.Repository) error {
		active, err := incidents.GetActiveByTopicID(ctx, chatID, topicID)
		if err != nil {
			return err
		}

		if active.Severity == severity {
			updated = active
			return nil
		}

		from := active.Severity
		if err := incidents.UpdateSeverity(ctx, active.ID, severity); err != nil {
			return err
		}

		if err := events.Create(ctx, &event.Event{
			IncidentID: active.ID,
			Type:       event.TypeSeverityChanged,
			UserID:     userID,
			Username:   username,
			Message:    fmt.Sprintf("%s → %s", from, severity),
		}); err != nil {
			return err
		}

		active.Severity = severity
		updated = active
		return nil
	})
	if err != nil {
		return nil, err
	}

	return updated, nil
}
