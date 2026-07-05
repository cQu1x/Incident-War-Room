package bot

import (
	"errors"
	"log"
	"strings"

	"gopkg.in/telebot.v3"

	"github.com/cQu1x/Incident-War-Room/internal/errs"
)

// HandleTopicText records a plain text message posted in an incident topic on
// the timeline. Media messages (screenshots, voice, etc.) are not handled here
// and never reach the timeline.
func (h *Handler) HandleTopicText(c telebot.Context) error {
	m := c.Message()
	if m == nil || m.Text == "" {
		return nil
	}
	return h.captureTopicMessage(c, m.Text)
}

// captureTopicMessage silently appends message to the active incident timeline
// for the current topic. Messages outside a topic, or in a topic without an
// active incident, are ignored.
func (h *Handler) captureTopicMessage(c telebot.Context, message string) error {
	message = strings.TrimSpace(message)
	if message == "" {
		return nil
	}

	topicID := threadID(c)
	if topicID == 0 {
		return nil
	}

	ctx, cancel := reqContext()
	defer cancel()

	userID, username := sender(c)
	if _, err := h.svc.AddTimelineEvent(ctx, c.Chat().ID, topicID, userID, username, message); err != nil {
		if errors.Is(err, errs.ErrNoActiveIncident) {
			return nil
		}
		log.Printf("bot: capture topic message: %v", err)
	}

	return nil
}
