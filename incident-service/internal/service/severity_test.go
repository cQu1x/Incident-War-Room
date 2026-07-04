package service

import (
	"context"
	"testing"

	"github.com/cQu1x/Incident-War-Room/internal/domain/event"
	"github.com/cQu1x/Incident-War-Room/internal/domain/incident"
	"github.com/cQu1x/Incident-War-Room/internal/errs"
)

func TestSetSeverity(t *testing.T) {
	ctx := context.Background()

	t.Run("updates the severity and records the transition", func(t *testing.T) {
		svc, incidents, events := newTestService()
		created, err := svc.CreateIncident(ctx, 200, 200, "DB is down", incident.SeverityLow, nil, "alice")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		inc, err := svc.SetSeverity(ctx, 200, 200, incident.SeverityHigh, ptrInt64(7), "bob")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if inc.Severity != incident.SeverityHigh {
			t.Fatalf("returned severity = %q, want HIGH", inc.Severity)
		}

		got, _ := incidents.GetActiveByTopicID(ctx, 200, 200)
		if got.Severity != incident.SeverityHigh {
			t.Fatalf("stored severity = %q, want HIGH", got.Severity)
		}

		evs := events.byIncident[created.ID]
		if len(evs) != 2 {
			t.Fatalf("expected creation + severity events, got %+v", evs)
		}
		last := evs[1]
		if last.Type != event.TypeSeverityChanged {
			t.Fatalf("expected SEVERITY_CHANGED event, got %q", last.Type)
		}
		if last.Message != "LOW → HIGH" {
			t.Fatalf("expected transition message, got %q", last.Message)
		}
		if last.Username != "bob" {
			t.Fatalf("expected severity change by bob, got %q", last.Username)
		}
	})

	t.Run("setting the same severity records no event", func(t *testing.T) {
		svc, _, events := newTestService()
		created, _ := svc.CreateIncident(ctx, 210, 210, "outage", incident.SeverityHigh, nil, "alice")

		if _, err := svc.SetSeverity(ctx, 210, 210, incident.SeverityHigh, nil, "bob"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if evs := events.byIncident[created.ID]; len(evs) != 1 {
			t.Fatalf("expected only the creation event, got %+v", evs)
		}
	})

	t.Run("invalid severity is rejected", func(t *testing.T) {
		svc, _, _ := newTestService()
		if _, err := svc.CreateIncident(ctx, 201, 201, "outage", incident.SeverityLow, nil, "alice"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		_, err := svc.SetSeverity(ctx, 201, 201, incident.Severity("CRITICAL"), nil, "bob")
		if errs.KindOf(err) != errs.KindValidation {
			t.Fatalf("expected validation error, got %v", err)
		}
	})

	t.Run("no active incident", func(t *testing.T) {
		svc, _, _ := newTestService()

		_, err := svc.SetSeverity(ctx, 999, 999, incident.SeverityHigh, nil, "bob")
		if errs.KindOf(err) != errs.KindNotFound {
			t.Fatalf("expected not-found, got %v", err)
		}
	})
}
