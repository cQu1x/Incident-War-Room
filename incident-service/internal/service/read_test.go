package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/cQu1x/Incident-War-Room/internal/domain/incident"
	"github.com/cQu1x/Incident-War-Room/internal/errs"
)

func TestGetIncident(t *testing.T) {
	ctx := context.Background()

	t.Run("returns the stored incident", func(t *testing.T) {
		svc, _, _ := newTestService()
		created, _ := svc.CreateIncident(ctx, 300, 300, "db down", incident.SeverityHigh, ptrInt64(1), "alice")

		got, err := svc.GetIncident(ctx, created.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ID != created.ID || got.Title != "db down" {
			t.Fatalf("unexpected incident: %+v", got)
		}
	})

	t.Run("unknown id is not found", func(t *testing.T) {
		svc, _, _ := newTestService()

		_, err := svc.GetIncident(ctx, uuid.New())
		if errs.KindOf(err) != errs.KindNotFound {
			t.Fatalf("expected not-found, got %v", err)
		}
	})
}

func TestListIncidents(t *testing.T) {
	ctx := context.Background()

	t.Run("empty when nothing was opened", func(t *testing.T) {
		svc, _, _ := newTestService()

		incidents, err := svc.ListIncidents(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(incidents) != 0 {
			t.Fatalf("expected no incidents, got %d", len(incidents))
		}
	})

	t.Run("returns every opened incident", func(t *testing.T) {
		svc, _, _ := newTestService()
		if _, err := svc.CreateIncident(ctx, 310, 310, "one", incident.SeverityLow, nil, "alice"); err != nil {
			t.Fatalf("create: %v", err)
		}
		if _, err := svc.CreateIncident(ctx, 311, 311, "two", incident.SeverityLow, nil, "bob"); err != nil {
			t.Fatalf("create: %v", err)
		}

		incidents, err := svc.ListIncidents(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(incidents) != 2 {
			t.Fatalf("expected 2 incidents, got %d", len(incidents))
		}
	})
}

func newReportService(reports *fakeReports) (*Service, *fakeIncidents) {
	incidents := newFakeIncidents()
	events := newFakeEvents()
	svc := New(incidents, events, fakeTx{incidents: incidents, events: events}, reports, &fakeTimelines{}, nil)
	return svc, incidents
}

func TestGenerateReport(t *testing.T) {
	ctx := context.Background()

	t.Run("persists the report url on the incident", func(t *testing.T) {
		reports := &fakeReports{url: "https://reports.example/r/abc.pdf"}
		svc, incidents := newReportService(reports)
		created, _ := svc.CreateIncident(ctx, 320, 320, "db down", incident.SeverityHigh, ptrInt64(1), "alice")

		doc, err := svc.GenerateReport(ctx, 320, 320)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if doc.URL != reports.url {
			t.Fatalf("doc url = %q, want %q", doc.URL, reports.url)
		}

		stored, _ := incidents.GetByID(ctx, created.ID)
		if stored.ReportURL == nil || *stored.ReportURL != reports.url {
			t.Fatalf("stored report url = %v, want %q", stored.ReportURL, reports.url)
		}
	})

	t.Run("propagates a generator failure", func(t *testing.T) {
		reports := &fakeReports{err: errors.New("report backend down")}
		svc, _ := newReportService(reports)
		if _, err := svc.CreateIncident(ctx, 321, 321, "db down", incident.SeverityHigh, nil, "alice"); err != nil {
			t.Fatalf("create: %v", err)
		}

		if _, err := svc.GenerateReport(ctx, 321, 321); err == nil {
			t.Fatal("expected error from failing report generator")
		}
	})

	t.Run("no active incident is not found", func(t *testing.T) {
		svc, _ := newReportService(&fakeReports{url: "https://reports.example/r/x.pdf"})

		if _, err := svc.GenerateReport(ctx, 322, 322); errs.KindOf(err) != errs.KindNotFound {
			t.Fatalf("expected not-found, got %v", err)
		}
	})
}
