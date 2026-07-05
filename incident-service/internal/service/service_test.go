package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/cQu1x/Incident-War-Room/internal/domain/event"
	"github.com/cQu1x/Incident-War-Room/internal/domain/incident"
	"github.com/cQu1x/Incident-War-Room/internal/domain/media"
	"github.com/cQu1x/Incident-War-Room/internal/domain/report"
	"github.com/cQu1x/Incident-War-Room/internal/domain/timeline"
	"github.com/cQu1x/Incident-War-Room/internal/errs"
)

type fakeIncidents struct {
	byID map[uuid.UUID]*incident.Incident
}

func newFakeIncidents() *fakeIncidents {
	return &fakeIncidents{byID: make(map[uuid.UUID]*incident.Incident)}
}

func (f *fakeIncidents) Create(_ context.Context, inc *incident.Incident) error {
	for _, existing := range f.byID {
		if existing.ChatID == inc.ChatID && existing.TopicID == inc.TopicID && existing.Status == incident.StatusActive {
			return errs.ErrIncidentAlreadyActive
		}
	}
	inc.ID = uuid.New()
	stored := *inc
	f.byID[inc.ID] = &stored
	return nil
}

func (f *fakeIncidents) GetByID(_ context.Context, id uuid.UUID) (*incident.Incident, error) {
	inc, ok := f.byID[id]
	if !ok {
		return nil, errs.ErrIncidentNotFound
	}
	clone := *inc
	return &clone, nil
}

func (f *fakeIncidents) List(_ context.Context, chatID *int64) ([]incident.Incident, error) {
	incidents := make([]incident.Incident, 0, len(f.byID))
	for _, inc := range f.byID {
		if chatID != nil && inc.ChatID != *chatID {
			continue
		}
		incidents = append(incidents, *inc)
	}
	return incidents, nil
}

func (f *fakeIncidents) GetActiveByTopicID(_ context.Context, chatID, topicID int64) (*incident.Incident, error) {
	for _, inc := range f.byID {
		if inc.ChatID == chatID && inc.TopicID == topicID && inc.Status == incident.StatusActive {
			clone := *inc
			return &clone, nil
		}
	}
	return nil, errs.ErrNoActiveIncident
}

func (f *fakeIncidents) CountActiveByTitle(_ context.Context, chatID int64, title string) (int, error) {
	count := 0
	for _, inc := range f.byID {
		if inc.ChatID == chatID && inc.Status == incident.StatusActive && inc.Title == title {
			count++
		}
	}
	return count, nil
}

func (f *fakeIncidents) UpdateSeverity(_ context.Context, id uuid.UUID, severity incident.Severity) error {
	inc, ok := f.byID[id]
	if !ok {
		return errs.ErrIncidentNotFound
	}
	inc.Severity = severity
	return nil
}

func (f *fakeIncidents) UpdateTelegraphURLs(_ context.Context, id uuid.UUID, telegraphURLs []string) error {
	inc, ok := f.byID[id]
	if !ok {
		return errs.ErrIncidentNotFound
	}
	inc.TelegraphURLs = telegraphURLs
	return nil
}

func (f *fakeIncidents) UpdateReportURL(_ context.Context, id uuid.UUID, reportURL string) error {
	inc, ok := f.byID[id]
	if !ok {
		return errs.ErrIncidentNotFound
	}
	inc.ReportURL = &reportURL
	return nil
}

func (f *fakeIncidents) Close(_ context.Context, id uuid.UUID, closedAt time.Time) error {
	inc, ok := f.byID[id]
	if !ok {
		return errs.ErrIncidentNotFound
	}
	if inc.Status != incident.StatusActive {
		return errs.ErrIncidentAlreadyClosed
	}
	inc.Status = incident.StatusClosed
	inc.ClosedAt = &closedAt
	return nil
}

func (f *fakeIncidents) Reopen(_ context.Context, id uuid.UUID, newTopicID int64) error {
	inc, ok := f.byID[id]
	if !ok {
		return errs.ErrIncidentNotFound
	}
	if inc.Status != incident.StatusClosed {
		return errs.ErrIncidentNotClosed
	}
	for _, other := range f.byID {
		if other.ID != id && other.ChatID == inc.ChatID && other.TopicID == newTopicID && other.Status == incident.StatusActive {
			return errs.ErrIncidentAlreadyActive
		}
	}
	inc.Status = incident.StatusActive
	inc.ClosedAt = nil
	inc.TopicID = newTopicID
	return nil
}

type fakeEvents struct {
	byIncident map[uuid.UUID][]event.Event
}

func newFakeEvents() *fakeEvents {
	return &fakeEvents{byIncident: make(map[uuid.UUID][]event.Event)}
}

func (f *fakeEvents) Create(_ context.Context, e *event.Event) error {
	e.ID = uuid.New()
	f.byIncident[e.IncidentID] = append(f.byIncident[e.IncidentID], *e)
	return nil
}

func (f *fakeEvents) ListByIncidentID(_ context.Context, incidentID uuid.UUID) ([]event.Event, error) {
	return f.byIncident[incidentID], nil
}

func (f *fakeEvents) ListParticipants(_ context.Context, incidentID uuid.UUID) ([]int64, error) {
	seen := make(map[int64]struct{})
	var ids []int64
	for _, e := range f.byIncident[incidentID] {
		if e.UserID == nil {
			continue
		}
		if _, ok := seen[*e.UserID]; ok {
			continue
		}
		seen[*e.UserID] = struct{}{}
		ids = append(ids, *e.UserID)
	}
	return ids, nil
}

type fakeTx struct {
	incidents incident.Repository
	events    event.Repository
}

func (f fakeTx) WithTx(_ context.Context, fn func(incident.Repository, event.Repository) error) error {
	return fn(f.incidents, f.events)
}

type fakeReports struct {
	last report.Report
	url  string
	err  error
}

func (f *fakeReports) Generate(_ context.Context, r report.Report) (report.Document, error) {
	f.last = r
	if f.err != nil {
		return report.Document{}, f.err
	}
	return report.Document{URL: f.url}, nil
}

type fakeTimelines struct {
	last timeline.Timeline
	urls []string
	err  error
}

func (f *fakeTimelines) Publish(_ context.Context, t timeline.Timeline) ([]string, error) {
	f.last = t
	if f.err != nil {
		return nil, f.err
	}
	return f.urls, nil
}

type fakeMedia struct {
	lastKey  string
	lastFile media.File
	uploads  int
	url      string
	err      error
}

func (f *fakeMedia) Upload(_ context.Context, key string, file media.File) (string, error) {
	f.lastKey = key
	f.lastFile = file
	f.uploads++
	if f.err != nil {
		return "", f.err
	}
	return f.url, nil
}

func newTestService() (*Service, *fakeIncidents, *fakeEvents) {
	incidents := newFakeIncidents()
	events := newFakeEvents()
	svc := New(incidents, events, fakeTx{incidents: incidents, events: events}, &fakeReports{}, &fakeTimelines{}, nil)
	return svc, incidents, events
}

func ptrInt64(v int64) *int64 { return &v }

func TestCreateIncident(t *testing.T) {
	ctx := context.Background()

	t.Run("success writes incident and creation event", func(t *testing.T) {
		svc, _, events := newTestService()

		inc, err := svc.CreateIncident(ctx, 100, 100, "DB is down", incident.SeverityHigh, ptrInt64(7), "alice")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if inc.ID == uuid.Nil {
			t.Fatal("expected generated incident ID")
		}
		if inc.Status != incident.StatusActive {
			t.Fatalf("expected ACTIVE status, got %q", inc.Status)
		}

		evs := events.byIncident[inc.ID]
		if len(evs) != 1 || evs[0].Type != event.TypeIncidentCreated {
			t.Fatalf("expected one INCIDENT_CREATED event, got %+v", evs)
		}
	})

	t.Run("empty severity defaults to MEDIUM", func(t *testing.T) {
		svc, _, _ := newTestService()

		inc, err := svc.CreateIncident(ctx, 101, 101, "Latency spike", "", nil, "bob")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if inc.Severity != incident.SeverityMedium {
			t.Fatalf("expected MEDIUM severity, got %q", inc.Severity)
		}
	})

	t.Run("duplicate active incident is a conflict", func(t *testing.T) {
		svc, _, _ := newTestService()

		if _, err := svc.CreateIncident(ctx, 102, 102, "first", incident.SeverityLow, nil, "alice"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		_, err := svc.CreateIncident(ctx, 102, 102, "second", incident.SeverityLow, nil, "alice")
		if errs.KindOf(err) != errs.KindConflict {
			t.Fatalf("expected conflict, got %v", err)
		}
	})

	t.Run("empty title is a validation error", func(t *testing.T) {
		svc, _, _ := newTestService()

		_, err := svc.CreateIncident(ctx, 103, 103, "   ", incident.SeverityLow, nil, "alice")
		if errs.KindOf(err) != errs.KindValidation {
			t.Fatalf("expected validation error, got %v", err)
		}
	})

	t.Run("duplicate active title in a chat gets a numeric suffix", func(t *testing.T) {
		svc, _, _ := newTestService()

		// Same chat, different topics (a forum chat can hold many active
		// incidents), all opened with the same title.
		first, err := svc.CreateIncident(ctx, 104, 1, "DB is down", incident.SeverityLow, nil, "alice")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		second, err := svc.CreateIncident(ctx, 104, 2, "DB is down", incident.SeverityLow, nil, "alice")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		third, err := svc.CreateIncident(ctx, 104, 3, "DB is down", incident.SeverityLow, nil, "alice")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if first.Title != "DB is down" {
			t.Fatalf("first title = %q, want %q", first.Title, "DB is down")
		}
		if second.Title != "DB is down-2" {
			t.Fatalf("second title = %q, want %q", second.Title, "DB is down-2")
		}
		if third.Title != "DB is down-3" {
			t.Fatalf("third title = %q, want %q", third.Title, "DB is down-3")
		}
	})

	t.Run("title is reused once the earlier incident is no longer active", func(t *testing.T) {
		svc, incidents, _ := newTestService()

		first, _ := svc.CreateIncident(ctx, 105, 1, "outage", incident.SeverityLow, nil, "alice")
		if err := incidents.Close(ctx, first.ID, time.Now()); err != nil {
			t.Fatalf("close: %v", err)
		}

		reused, err := svc.CreateIncident(ctx, 105, 2, "outage", incident.SeverityLow, nil, "alice")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if reused.Title != "outage" {
			t.Fatalf("title = %q, want %q (base is free again)", reused.Title, "outage")
		}
	})
}

func TestAddTimelineEvent(t *testing.T) {
	ctx := context.Background()

	t.Run("appends a comment to the active incident", func(t *testing.T) {
		svc, _, events := newTestService()
		inc, _ := svc.CreateIncident(ctx, 200, 200, "outage", incident.SeverityHigh, nil, "alice")

		e, err := svc.AddTimelineEvent(ctx, 200, 200, ptrInt64(9), "bob", "investigating")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if e.Type != event.TypeCommentAdded || e.IncidentID != inc.ID {
			t.Fatalf("unexpected event: %+v", e)
		}
		if len(events.byIncident[inc.ID]) != 2 {
			t.Fatalf("expected 2 events (created + comment), got %d", len(events.byIncident[inc.ID]))
		}
	})

	t.Run("no active incident", func(t *testing.T) {
		svc, _, _ := newTestService()

		_, err := svc.AddTimelineEvent(ctx, 201, 201, nil, "bob", "hello")
		if errs.KindOf(err) != errs.KindNotFound {
			t.Fatalf("expected not-found, got %v", err)
		}
	})

	t.Run("empty message is a validation error", func(t *testing.T) {
		svc, _, _ := newTestService()
		_, _ = svc.CreateIncident(ctx, 202, 202, "outage", incident.SeverityHigh, nil, "alice")

		_, err := svc.AddTimelineEvent(ctx, 202, 202, nil, "bob", "  ")
		if errs.KindOf(err) != errs.KindValidation {
			t.Fatalf("expected validation error, got %v", err)
		}
	})
}

func TestAddTimelineEventWithMedia(t *testing.T) {
	ctx := context.Background()

	newImageService := func(m media.Storage) (*Service, *fakeIncidents, *fakeEvents) {
		incidents := newFakeIncidents()
		events := newFakeEvents()
		svc := New(incidents, events, fakeTx{incidents: incidents, events: events}, &fakeReports{}, &fakeTimelines{}, m)
		return svc, incidents, events
	}

	t.Run("uploads media and stores its url on the event", func(t *testing.T) {
		store := &fakeMedia{url: "https://cdn.example/incidents/pic.jpg"}
		svc, _, events := newImageService(store)
		inc, _ := svc.CreateIncident(ctx, 500, 500, "outage", incident.SeverityHigh, nil, "alice")

		e, err := svc.AddTimelineEventWithMedia(ctx, 500, 500, ptrInt64(3), "bob", "screenshot", []media.File{{Data: []byte("x"), ContentType: "image/jpeg", Ext: "jpg"}})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(e.MediaURLs) != 1 || e.MediaURLs[0] != store.url {
			t.Fatalf("expected media url %q, got %v", store.url, e.MediaURLs)
		}
		if e.Message != "screenshot" {
			t.Fatalf("expected caption stored as message, got %q", e.Message)
		}
		stored := events.byIncident[inc.ID]
		if len(stored) != 2 || len(stored[1].MediaURLs) != 1 {
			t.Fatalf("expected comment event with media url, got %+v", stored)
		}
	})

	t.Run("uploads every file of an album onto one event", func(t *testing.T) {
		store := &fakeMedia{url: "https://cdn.example/incidents/pic.jpg"}
		svc, _, events := newImageService(store)
		inc, _ := svc.CreateIncident(ctx, 503, 503, "outage", incident.SeverityHigh, nil, "alice")

		files := []media.File{{Ext: "jpg"}, {Ext: "mp4"}, {Ext: "png"}}
		e, err := svc.AddTimelineEventWithMedia(ctx, 503, 503, ptrInt64(3), "bob", "album", files)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(e.MediaURLs) != 3 {
			t.Fatalf("expected 3 media urls, got %v", e.MediaURLs)
		}
		if store.uploads != 3 {
			t.Fatalf("expected 3 uploads, got %d", store.uploads)
		}
		stored := events.byIncident[inc.ID]
		if len(stored) != 2 || len(stored[1].MediaURLs) != 3 {
			t.Fatalf("expected comment event with 3 media urls, got %+v", stored)
		}
	})

	t.Run("no files is a validation error", func(t *testing.T) {
		svc, _, _ := newImageService(&fakeMedia{url: "https://cdn.example/x.jpg"})
		_, _ = svc.CreateIncident(ctx, 504, 504, "outage", incident.SeverityHigh, nil, "alice")

		_, err := svc.AddTimelineEventWithMedia(ctx, 504, 504, nil, "bob", "", nil)
		if errs.KindOf(err) != errs.KindValidation {
			t.Fatalf("expected validation, got %v", err)
		}
	})

	t.Run("no media storage is unavailable", func(t *testing.T) {
		svc, _, _ := newImageService(nil)
		_, _ = svc.CreateIncident(ctx, 501, 501, "outage", incident.SeverityHigh, nil, "alice")

		_, err := svc.AddTimelineEventWithMedia(ctx, 501, 501, nil, "bob", "", []media.File{{Ext: "jpg"}})
		if errs.KindOf(err) != errs.KindUnavailable {
			t.Fatalf("expected unavailable, got %v", err)
		}
	})

	t.Run("no active incident", func(t *testing.T) {
		svc, _, _ := newImageService(&fakeMedia{url: "https://cdn.example/x.jpg"})

		_, err := svc.AddTimelineEventWithMedia(ctx, 502, 502, nil, "bob", "", []media.File{{Ext: "jpg"}})
		if errs.KindOf(err) != errs.KindNotFound {
			t.Fatalf("expected not-found, got %v", err)
		}
	})
}

func TestGetTimeline(t *testing.T) {
	ctx := context.Background()

	t.Run("returns incident and its events in order", func(t *testing.T) {
		svc, _, _ := newTestService()
		inc, _ := svc.CreateIncident(ctx, 400, 400, "outage", incident.SeverityHigh, nil, "alice")
		_, _ = svc.AddTimelineEvent(ctx, 400, 400, ptrInt64(1), "bob", "looking into it")

		got, events, err := svc.GetTimeline(ctx, 400, 400)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ID != inc.ID {
			t.Fatalf("expected incident %s, got %s", inc.ID, got.ID)
		}
		if len(events) != 2 {
			t.Fatalf("expected 2 events, got %d", len(events))
		}
		if events[0].Type != event.TypeIncidentCreated || events[1].Type != event.TypeCommentAdded {
			t.Fatalf("unexpected event order: %+v", events)
		}
	})

	t.Run("no active incident", func(t *testing.T) {
		svc, _, _ := newTestService()

		_, _, err := svc.GetTimeline(ctx, 401, 401)
		if errs.KindOf(err) != errs.KindNotFound {
			t.Fatalf("expected not-found, got %v", err)
		}
	})
}

func TestPublishTimeline(t *testing.T) {
	ctx := context.Background()

	t.Run("publishes the active incident timeline and returns urls", func(t *testing.T) {
		incidents := newFakeIncidents()
		events := newFakeEvents()
		publisher := &fakeTimelines{urls: []string{"https://telegra.ph/timeline-1"}}
		svc := New(incidents, events, fakeTx{incidents: incidents, events: events}, &fakeReports{}, publisher, nil)

		inc, _ := svc.CreateIncident(ctx, 600, 600, "outage", incident.SeverityHigh, ptrInt64(1), "alice")
		_, _ = svc.AddTimelineEvent(ctx, 600, 600, ptrInt64(1), "alice", "looking into it")

		urls, err := svc.PublishTimeline(ctx, 600, 600)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(urls) != 1 || urls[0] != "https://telegra.ph/timeline-1" {
			t.Fatalf("unexpected urls: %v", urls)
		}
		if publisher.last.Incident.ID != inc.ID || len(publisher.last.Events) != 2 {
			t.Fatalf("publisher received %+v", publisher.last)
		}

		stored, _ := incidents.GetByID(ctx, inc.ID)
		if len(stored.TelegraphURLs) != 1 || stored.TelegraphURLs[0] != "https://telegra.ph/timeline-1" {
			t.Fatalf("expected telegraph urls persisted, got %v", stored.TelegraphURLs)
		}
	})

	t.Run("no active incident", func(t *testing.T) {
		svc, _, _ := newTestService()

		_, err := svc.PublishTimeline(ctx, 601, 601)
		if errs.KindOf(err) != errs.KindNotFound {
			t.Fatalf("expected not-found, got %v", err)
		}
	})
}

func TestCloseIncident(t *testing.T) {
	ctx := context.Background()

	t.Run("closes incident and writes close event", func(t *testing.T) {
		svc, _, events := newTestService()
		inc, _ := svc.CreateIncident(ctx, 500, 500, "outage", incident.SeverityHigh, nil, "alice")

		closed, err := svc.CloseIncident(ctx, 500, 500, ptrInt64(3), "carol")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if closed.Status != incident.StatusClosed {
			t.Fatalf("expected CLOSED status, got %q", closed.Status)
		}
		if closed.ClosedAt == nil {
			t.Fatal("expected ClosedAt to be set")
		}

		evs := events.byIncident[inc.ID]
		if len(evs) != 2 || evs[1].Type != event.TypeIncidentClosed {
			t.Fatalf("expected creation + close events, got %+v", evs)
		}
	})

	t.Run("no active incident", func(t *testing.T) {
		svc, _, _ := newTestService()

		_, err := svc.CloseIncident(ctx, 501, 501, nil, "carol")
		if errs.KindOf(err) != errs.KindNotFound {
			t.Fatalf("expected not-found, got %v", err)
		}
	})

	t.Run("closing twice is a conflict", func(t *testing.T) {
		svc, _, _ := newTestService()
		_, _ = svc.CreateIncident(ctx, 502, 502, "outage", incident.SeverityHigh, nil, "alice")

		if _, err := svc.CloseIncident(ctx, 502, 502, nil, "carol"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		_, err := svc.CloseIncident(ctx, 502, 502, nil, "carol")
		if errs.KindOf(err) != errs.KindNotFound {
			t.Fatalf("expected not-found, got %v", err)
		}
	})
}
