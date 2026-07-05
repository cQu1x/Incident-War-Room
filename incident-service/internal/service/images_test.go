package service

import (
	"context"
	"testing"

	"github.com/cQu1x/Incident-War-Room/internal/domain/incident"
	"github.com/cQu1x/Incident-War-Room/internal/domain/media"
	"github.com/cQu1x/Incident-War-Room/internal/errs"
)

func newImageTestService(store media.Storage) *Service {
	incidents := newFakeIncidents()
	events := newFakeEvents()
	return New(incidents, events, fakeTx{incidents: incidents, events: events}, &fakeReports{}, &fakeTimelines{}, store)
}

func TestIncidentImages(t *testing.T) {
	ctx := context.Background()

	t.Run("returns only events that carry an image", func(t *testing.T) {
		store := &fakeMedia{url: "https://cdn.example/pic.jpg"}
		svc := newImageTestService(store)
		inc, _ := svc.CreateIncident(ctx, 700, 700, "outage", incident.SeverityHigh, ptrInt64(1), "alice")
		if _, err := svc.AddTimelineEvent(ctx, 700, 700, ptrInt64(2), "bob", "looking"); err != nil {
			t.Fatalf("add event: %v", err)
		}
		if _, err := svc.AddTimelineEventWithMedia(ctx, 700, 700, ptrInt64(3), "carol", "screenshot", []media.File{{Data: []byte("x"), Ext: "jpg"}}); err != nil {
			t.Fatalf("add image: %v", err)
		}

		images, err := svc.IncidentImages(ctx, inc.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(images) != 1 {
			t.Fatalf("expected 1 image event, got %d", len(images))
		}
		if len(images[0].MediaURLs) != 1 || images[0].MediaURLs[0] != store.url {
			t.Fatalf("expected media url %q, got %v", store.url, images[0].MediaURLs)
		}
	})

	t.Run("unknown incident is not found", func(t *testing.T) {
		svc := newImageTestService(&fakeMedia{})
		if _, err := svc.IncidentImages(ctx, incident.Incident{}.ID); errs.KindOf(err) != errs.KindNotFound {
			t.Fatalf("expected not-found, got %v", err)
		}
	})
}

func TestIncidentTimelineReturnsIncidentAndEvents(t *testing.T) {
	ctx := context.Background()
	svc, _, _ := newTestService()

	inc, _ := svc.CreateIncident(ctx, 800, 800, "db down", incident.SeverityHigh, ptrInt64(1), "alice")
	if _, err := svc.AddTimelineEvent(ctx, 800, 800, ptrInt64(2), "bob", "investigating"); err != nil {
		t.Fatalf("add event: %v", err)
	}

	got, events, err := svc.IncidentTimeline(ctx, inc.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != inc.ID || got.Title != "db down" {
		t.Fatalf("unexpected incident: %+v", got)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
}
