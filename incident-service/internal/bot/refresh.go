package bot

import (
	"context"
	"errors"
	"log"

	"github.com/cQu1x/Incident-War-Room/internal/errs"
)

// timelineJob tracks the background Telegraph refresh for one incident topic.
// active marks a running publish; again coalesces any requests that arrive while
// it runs into a single follow-up pass.
type timelineJob struct {
	active bool
	again  bool
}

// refreshTimeline republishes the topic's Telegraph timeline in the background
// so the page stays live as events are added. Refreshes for the same topic are
// serialized and coalesced, so a burst of messages never races two publishes or
// piles up goroutines.
func (h *Handler) refreshTimeline(chatID, topicID int64) {
	if topicID == 0 {
		return
	}

	key := announceKey{chatID, topicID}

	h.timelineMu.Lock()
	job := h.timelineJobs[key]
	if job == nil {
		job = &timelineJob{}
		h.timelineJobs[key] = job
	}
	if job.active {
		job.again = true
		h.timelineMu.Unlock()
		return
	}
	job.active = true
	h.timelineMu.Unlock()

	go h.runTimelineRefresh(key)
}

func (h *Handler) runTimelineRefresh(key announceKey) {
	for {
		ctx, cancel := context.WithTimeout(context.Background(), handlerTimeout)
		_, err := h.svc.PublishTimeline(ctx, key.chatID, key.topicID)
		cancel()
		if err != nil && !errors.Is(err, errs.ErrNoActiveIncident) {
			log.Printf("bot: refresh timeline for topic %d: %v", key.topicID, err)
		}

		h.timelineMu.Lock()
		job := h.timelineJobs[key]
		if job != nil && job.again {
			job.again = false
			h.timelineMu.Unlock()
			continue
		}
		delete(h.timelineJobs, key)
		h.timelineMu.Unlock()
		return
	}
}
