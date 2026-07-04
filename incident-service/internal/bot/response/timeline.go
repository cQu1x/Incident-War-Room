package response

import (
	"fmt"
	"strings"

	"github.com/cQu1x/Incident-War-Room/internal/domain/event"
	"github.com/cQu1x/Incident-War-Room/internal/domain/incident"
)

const maxInlineEvents = 5

func Timeline(inc incident.Incident, events []event.Event) string {
	var b strings.Builder

	fmt.Fprintf(&b, "📋 <b>Timeline</b> — %s\n", escape(inc.Title))

	if len(events) == 0 {
		b.WriteString("\nThe incident timeline is empty.")
		return b.String()
	}

	shown := events
	if len(events) > maxInlineEvents {
		shown = events[len(events)-maxInlineEvents:]
		fmt.Fprintf(&b, "\n<i>Showing the last %d of %d updates.</i>\n", maxInlineEvents, len(events))
	}

	for _, e := range shown {
		line := escape(e.Username)
		if summary := e.Summary(); summary != "" {
			line += ": " + escape(summary)
		}
		fmt.Fprintf(&b, "\n<b>%s</b> — %s", formatTime(e.CreatedAt), line)
	}

	return b.String()
}

func TimelineUnavailable() string {
	return "\n\n📄 <i>Telegraph is unavailable right now — the full timeline was not published.</i>"
}

func TimelineLink(urls []string) string {
	if len(urls) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("\n\n📄 <b>Full timeline:</b>")
	if len(urls) == 1 {
		fmt.Fprintf(&b, " <a href=\"%s\">Telegraph</a>", escape(urls[0]))
		return b.String()
	}
	for i, u := range urls {
		fmt.Fprintf(&b, "\n<a href=\"%s\">Part %d</a>", escape(u), i+1)
	}
	return b.String()
}
