package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/cQu1x/Incident-War-Room/internal/domain/event"
	"github.com/cQu1x/Incident-War-Room/internal/domain/incident"
	"github.com/cQu1x/Incident-War-Room/internal/domain/media"
	"github.com/cQu1x/Incident-War-Room/internal/errs"
)

// CreateIncident opens a new incident in the chat and records an
// INCIDENT_CREATED event on its timeline. Both writes happen in one
// transaction.
//
// Returns errs.ErrIncidentAlreadyActive if the chat already has an active
// incident, or an errs.KindValidation error if the input is invalid. An empty
// severity defaults to incident.SeverityMedium.
func (s *Service) CreateIncident(
	ctx context.Context,
	chatID int64,
	topicID int64,
	title string,
	severity incident.Severity,
	userID *int64,
	username string,
) (*incident.Incident, error) {
	const op = "service.CreateIncident"

	title = strings.TrimSpace(title)
	if title == "" {
		return nil, errs.New(errs.KindValidation, op, "incident title is required")
	}

	if severity == "" {
		severity = incident.SeverityMedium
	}
	if !validSeverity(severity) {
		return nil, errs.New(errs.KindValidation, op, "invalid severity")
	}

	inc := &incident.Incident{
		Severity:  severity,
		Status:    incident.StatusActive,
		ChatID:    chatID,
		TopicID:   topicID,
		CreatedBy: userID,
	}

	err := s.tx.WithTx(ctx, func(incidents incident.Repository, events event.Repository) error {
		uniqueTitle, err := uniqueActiveTitle(ctx, incidents, chatID, title)
		if err != nil {
			return err
		}
		inc.Title = uniqueTitle

		if err := incidents.Create(ctx, inc); err != nil {
			return err
		}
		return events.Create(ctx, &event.Event{
			IncidentID: inc.ID,
			Type:       event.TypeIncidentCreated,
			UserID:     userID,
			Username:   username,
			Message:    inc.Title,
		})
	})
	if err != nil {
		return nil, err
	}

	return inc, nil
}

// AddTimelineEvent appends a comment to the chat's active incident timeline.
//
// Returns errs.ErrNoActiveIncident if the chat has no active incident, or an
// errs.KindValidation error if the message is empty.
func (s *Service) AddTimelineEvent(
	ctx context.Context,
	chatID int64,
	topicID int64,
	userID *int64,
	username string,
	message string,
) (*event.Event, error) {
	const op = "service.AddTimelineEvent"

	message = strings.TrimSpace(message)
	if message == "" {
		return nil, errs.New(errs.KindValidation, op, "message is required")
	}

	active, err := s.incidents.GetActiveByTopicID(ctx, chatID, topicID)
	if err != nil {
		return nil, err
	}

	e := &event.Event{
		IncidentID: active.ID,
		Type:       event.TypeCommentAdded,
		UserID:     userID,
		Username:   username,
		Message:    message,
	}
	if err := s.events.Create(ctx, e); err != nil {
		return nil, err
	}

	return e, nil
}

// AddTimelineEventWithMedia appends a comment carrying one or more media
// attachments (images, video, documents, …) to the chat's active incident
// timeline. Each file is uploaded to media storage and its public URL is stored
// on the event in MediaURLs, alongside the (possibly empty) caption.
//
// Returns errs.ErrNoActiveIncident if the chat has no active incident, an
// errs.KindUnavailable error if media storage is not configured, or an
// errs.KindValidation error if no files are provided.
func (s *Service) AddTimelineEventWithMedia(
	ctx context.Context,
	chatID int64,
	topicID int64,
	userID *int64,
	username string,
	caption string,
	files []media.File,
) (*event.Event, error) {
	const op = "service.AddTimelineEventWithMedia"

	if s.media == nil {
		return nil, errs.New(errs.KindUnavailable, op, "media storage is not configured")
	}
	if len(files) == 0 {
		return nil, errs.New(errs.KindValidation, op, "at least one media file is required")
	}

	active, err := s.incidents.GetActiveByTopicID(ctx, chatID, topicID)
	if err != nil {
		return nil, err
	}

	urls := make([]string, 0, len(files))
	for _, f := range files {
		key := fmt.Sprintf("incidents/%s/%s.%s", active.ID, uuid.New(), f.Ext)
		url, err := s.media.Upload(ctx, key, f)
		if err != nil {
			return nil, err
		}
		urls = append(urls, url)
	}

	e := &event.Event{
		IncidentID: active.ID,
		Type:       event.TypeCommentAdded,
		UserID:     userID,
		Username:   username,
		Message:    strings.TrimSpace(caption),
		MediaURLs:  urls,
	}
	if err := s.events.Create(ctx, e); err != nil {
		return nil, err
	}

	return e, nil
}

// uniqueActiveTitle returns a title that no other active incident in the chat
// currently uses. If base is free it is returned unchanged; otherwise a numeric
// suffix ("-2", "-3", …) is appended until a free title is found.
func uniqueActiveTitle(ctx context.Context, incidents incident.Repository, chatID int64, base string) (string, error) {
	candidate := base
	for n := 2; ; n++ {
		count, err := incidents.CountActiveByTitle(ctx, chatID, candidate)
		if err != nil {
			return "", err
		}
		if count == 0 {
			return candidate, nil
		}
		candidate = fmt.Sprintf("%s-%d", base, n)
	}
}

func validSeverity(s incident.Severity) bool {
	switch s {
	case incident.SeverityLow, incident.SeverityMedium, incident.SeverityHigh:
		return true
	default:
		return false
	}
}
