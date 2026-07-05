package api

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/cQu1x/Incident-War-Room/internal/domain/event"
	"github.com/cQu1x/Incident-War-Room/internal/domain/incident"
)

func strptr(s string) *string { return &s }

func TestNewTimelineResponseClosedHasDuration(t *testing.T) {
	start := time.Date(2026, 6, 26, 10, 0, 0, 0, time.UTC)
	end := start.Add(90 * time.Minute)
	inc := incident.Incident{
		ID:        uuid.New(),
		Title:     "db down",
		Status:    incident.StatusClosed,
		Severity:  incident.SeverityHigh,
		CreatedAt: start,
		ClosedAt:  &end,
	}
	events := []event.Event{
		{ID: uuid.New(), Username: "alice"},
		{ID: uuid.New(), Username: "bob"},
		{ID: uuid.New(), Username: "alice"},
		{ID: uuid.New(), Username: ""},
	}

	resp := newTimelineResponse(inc, events)

	if resp.DurationSeconds == nil || *resp.DurationSeconds != 5400 {
		t.Fatalf("duration = %v, want 5400", resp.DurationSeconds)
	}
	if resp.EndedAt == nil || !resp.EndedAt.Equal(end) {
		t.Fatalf("endedAt = %v, want %v", resp.EndedAt, end)
	}
	if got := resp.Responders; len(got) != 2 || got[0] != "alice" || got[1] != "bob" {
		t.Fatalf("responders = %v, want [alice bob]", got)
	}
	if resp.Title != "db down" || resp.Status != "CLOSED" {
		t.Fatalf("unexpected header: %+v", resp)
	}
}

func TestNewTimelineResponseActiveHasNoDuration(t *testing.T) {
	inc := incident.Incident{
		ID:        uuid.New(),
		Status:    incident.StatusActive,
		CreatedAt: time.Now(),
	}

	resp := newTimelineResponse(inc, nil)

	if resp.DurationSeconds != nil {
		t.Fatalf("duration = %v, want nil for active incident", *resp.DurationSeconds)
	}
	if resp.EndedAt != nil {
		t.Fatalf("endedAt = %v, want nil for active incident", resp.EndedAt)
	}
	if len(resp.Responders) != 0 {
		t.Fatalf("responders = %v, want empty", resp.Responders)
	}
}

func TestNewImageResponsesFiltersEventsWithoutMedia(t *testing.T) {
	events := []event.Event{
		{ID: uuid.New(), Username: "alice", Message: "no image"},
		{ID: uuid.New(), Username: "bob", Message: "shot", MediaURLs: []string{"https://cdn.example/a.jpg"}},
		{ID: uuid.New(), Username: "carol", Message: "album", MediaURLs: []string{"https://cdn.example/b.jpg", "https://cdn.example/c.mp4"}},
	}

	images := newImageResponses(events)

	if len(images) != 3 {
		t.Fatalf("expected 3 media items, got %d", len(images))
	}
	if images[0].URL != "https://cdn.example/a.jpg" || images[0].Username != "bob" || images[0].Message != "shot" {
		t.Fatalf("unexpected image: %+v", images[0])
	}
}
