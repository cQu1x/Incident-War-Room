package api

import (
	"time"

	"github.com/cQu1x/Incident-War-Room/internal/domain/event"
)

type eventResponse struct {
	ID         string    `json:"id"`
	IncidentID string    `json:"incidentId"`
	Type       string    `json:"type"`
	UserID     *int64    `json:"userId"`
	Username   string    `json:"username"`
	Message    string    `json:"message"`
	MediaURLs  []string  `json:"mediaUrls"`
	CreatedAt  time.Time `json:"createdAt"`
}

func newEventResponse(e event.Event) eventResponse {
	mediaURLs := e.MediaURLs
	if mediaURLs == nil {
		mediaURLs = []string{}
	}
	return eventResponse{
		ID:         e.ID.String(),
		IncidentID: e.IncidentID.String(),
		Type:       string(e.Type),
		UserID:     e.UserID,
		Username:   e.Username,
		Message:    e.Message,
		MediaURLs:  mediaURLs,
		CreatedAt:  e.CreatedAt,
	}
}

func newEventResponses(events []event.Event) []eventResponse {
	out := make([]eventResponse, 0, len(events))
	for _, e := range events {
		out = append(out, newEventResponse(e))
	}
	return out
}
