package service

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/cQu1x/Incident-War-Room/internal/domain/event"
	"github.com/cQu1x/Incident-War-Room/internal/domain/incident"
	"github.com/cQu1x/Incident-War-Room/internal/errs"
)

func TestReopenIncident(t *testing.T) {
	ctx := context.Background()

	t.Run("reopens closed incident under a new topic and keeps the timeline", func(t *testing.T) {
		svc, _, events := newTestService()
		inc, _ := svc.CreateIncident(ctx, 700, 700, "outage", incident.SeverityHigh, ptrInt64(1), "alice")
		_, _ = svc.AddTimelineEvent(ctx, 700, 700, ptrInt64(1), "alice", "looking into it")
		if _, err := svc.CloseIncident(ctx, 700, 700, ptrInt64(1), "alice"); err != nil {
			t.Fatalf("close: %v", err)
		}

		reopened, err := svc.ReopenIncident(ctx, inc.ID, 701, ptrInt64(2), "bob")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if reopened.Status != incident.StatusActive {
			t.Fatalf("expected ACTIVE status, got %q", reopened.Status)
		}
		if reopened.ClosedAt != nil {
			t.Fatalf("expected ClosedAt cleared, got %v", reopened.ClosedAt)
		}
		if reopened.TopicID != 701 {
			t.Fatalf("expected new topic 701, got %d", reopened.TopicID)
		}

		evs := events.byIncident[inc.ID]
		if len(evs) != 4 {
			t.Fatalf("expected created+comment+closed+reopened events, got %d: %+v", len(evs), evs)
		}
		if evs[3].Type != event.TypeIncidentReopened {
			t.Fatalf("expected last event INCIDENT_REOPENED, got %q", evs[3].Type)
		}
		if evs[3].Username != "bob" {
			t.Fatalf("expected reopen event by bob, got %q", evs[3].Username)
		}
	})

	t.Run("the incident stays reachable in its new topic", func(t *testing.T) {
		svc, _, _ := newTestService()
		inc, _ := svc.CreateIncident(ctx, 710, 710, "outage", incident.SeverityHigh, nil, "alice")
		_, _ = svc.CloseIncident(ctx, 710, 710, nil, "alice")

		if _, err := svc.ReopenIncident(ctx, inc.ID, 711, nil, "bob"); err != nil {
			t.Fatalf("reopen: %v", err)
		}

		got, _, err := svc.GetTimeline(ctx, 710, 711)
		if err != nil {
			t.Fatalf("expected the reopened incident to be active in topic 711: %v", err)
		}
		if got.ID != inc.ID {
			t.Fatalf("expected incident %s, got %s", inc.ID, got.ID)
		}
	})

	t.Run("reopening an active incident is a conflict", func(t *testing.T) {
		svc, _, _ := newTestService()
		inc, _ := svc.CreateIncident(ctx, 720, 720, "outage", incident.SeverityHigh, nil, "alice")

		_, err := svc.ReopenIncident(ctx, inc.ID, 721, nil, "bob")
		if errs.KindOf(err) != errs.KindConflict {
			t.Fatalf("expected conflict, got %v", err)
		}
	})

	t.Run("reopening a missing incident is not found", func(t *testing.T) {
		svc, _, _ := newTestService()

		_, err := svc.ReopenIncident(ctx, uuid.New(), 731, nil, "bob")
		if errs.KindOf(err) != errs.KindNotFound {
			t.Fatalf("expected not-found, got %v", err)
		}
	})
}
