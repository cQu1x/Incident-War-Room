package api

import (
	"time"

	"github.com/cQu1x/Incident-War-Room/internal/domain/event"
)

// imageResponse is a single media attachment on the incident timeline, with the
// caption it was posted with and its author. An event carrying several
// attachments yields one imageResponse per attachment.
type imageResponse struct {
	EventID   string    `json:"eventId"`
	URL       string    `json:"url"`
	Username  string    `json:"username"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"createdAt"`
}

// newImageResponses flattens timeline events into media DTOs, emitting one entry
// per attached media URL so every attachment of every event is surfaced.
func newImageResponses(events []event.Event) []imageResponse {
	out := make([]imageResponse, 0, len(events))
	for _, e := range events {
		for _, url := range e.MediaURLs {
			out = append(out, imageResponse{
				EventID:   e.ID.String(),
				URL:       url,
				Username:  e.Username,
				Message:   e.Message,
				CreatedAt: e.CreatedAt,
			})
		}
	}
	return out
}
