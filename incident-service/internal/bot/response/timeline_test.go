package response

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/cQu1x/Incident-War-Room/internal/domain/event"
	"github.com/cQu1x/Incident-War-Room/internal/domain/incident"
)

func TestTimelineEmpty(t *testing.T) {
	got := Timeline(incident.Incident{Title: "DB is down"}, nil)

	if !strings.Contains(got, "timeline is empty") {
		t.Errorf("Timeline() = %q, expected empty notice", got)
	}
	if !strings.Contains(got, "DB is down") {
		t.Errorf("Timeline() = %q, missing incident title", got)
	}
}

func TestTimelineWithEvents(t *testing.T) {
	inc := incident.Incident{Title: "DB is down"}
	events := []event.Event{
		{
			Type:      event.TypeIncidentCreated,
			Username:  "alice",
			Message:   "Incident opened",
			CreatedAt: time.Date(2026, 6, 13, 10, 0, 0, 0, time.UTC),
		},
		{
			Type:      event.TypeCommentAdded,
			Username:  "bob",
			Message:   "Restarting the primary",
			CreatedAt: time.Date(2026, 6, 13, 10, 5, 0, 0, time.UTC),
		},
	}

	got := Timeline(inc, events)

	for _, want := range []string{
		"alice", "Incident opened", "2026-06-13 10:00 UTC",
		"bob", "Restarting the primary", "2026-06-13 10:05 UTC",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("Timeline() = %q, missing %q", got, want)
		}
	}
	if strings.Contains(got, "timeline is empty") {
		t.Errorf("Timeline() with events should not show empty notice: %q", got)
	}
}

func TestTimelineLabelsLifecycleEvents(t *testing.T) {
	inc := incident.Incident{Title: "DB is down"}
	events := []event.Event{
		{Type: event.TypeIncidentCreated, Username: "alice", Message: "DB is down", CreatedAt: time.Date(2026, 6, 13, 10, 0, 0, 0, time.UTC)},
		{Type: event.TypeIncidentClosed, Username: "bob", CreatedAt: time.Date(2026, 6, 13, 11, 0, 0, 0, time.UTC)},
		{Type: event.TypeIncidentReopened, Username: "carol", CreatedAt: time.Date(2026, 6, 13, 12, 0, 0, 0, time.UTC)},
	}

	got := Timeline(inc, events)

	for _, want := range []string{"Incident opened: DB is down", "Incident closed", "Incident reopened"} {
		if !strings.Contains(got, want) {
			t.Errorf("Timeline() = %q, missing lifecycle label %q", got, want)
		}
	}
}

func TestTimelineShowsOnlyLastFive(t *testing.T) {
	base := time.Date(2026, 6, 13, 10, 0, 0, 0, time.UTC)
	events := make([]event.Event, 8)
	for i := range events {
		events[i] = event.Event{
			Username:  "alice",
			Message:   fmt.Sprintf("update-%d", i+1),
			CreatedAt: base.Add(time.Duration(i) * time.Minute),
		}
	}

	got := Timeline(incident.Incident{Title: "outage"}, events)

	for _, want := range []string{"update-4", "update-5", "update-6", "update-7", "update-8"} {
		if !strings.Contains(got, want) {
			t.Errorf("Timeline() = %q, missing recent %q", got, want)
		}
	}
	for _, gone := range []string{"update-1", "update-2", "update-3"} {
		if strings.Contains(got, gone) {
			t.Errorf("Timeline() = %q, should not show older %q", got, gone)
		}
	}
	if !strings.Contains(got, "last 5 of 8") {
		t.Errorf("Timeline() = %q, missing truncation notice", got)
	}
}

func TestTimelineEscapesMessage(t *testing.T) {
	events := []event.Event{
		{Username: "<b>eve</b>", Message: "<i>boom</i>", CreatedAt: time.Now()},
	}

	got := Timeline(incident.Incident{Title: "x"}, events)

	if strings.Contains(got, "<i>boom</i>") || strings.Contains(got, "<b>eve</b>") {
		t.Errorf("Timeline() did not escape event fields: %q", got)
	}
}
