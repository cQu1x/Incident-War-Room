package service

import (
	"context"
	"testing"

	"github.com/cQu1x/Incident-War-Room/internal/domain/incident"
)

func TestListIncidents(t *testing.T) {
	ctx := context.Background()

	setup := func() *Service {
		svc, _, _ := newTestService()
		if _, err := svc.CreateIncident(ctx, 100, 1, "DB down", incident.SeverityHigh, nil, "alice"); err != nil {
			t.Fatalf("seed chat 100 topic 1: %v", err)
		}
		if _, err := svc.CreateIncident(ctx, 100, 2, "Cache down", incident.SeverityLow, nil, "bob"); err != nil {
			t.Fatalf("seed chat 100 topic 2: %v", err)
		}
		if _, err := svc.CreateIncident(ctx, 200, 1, "API down", incident.SeverityMedium, nil, "carol"); err != nil {
			t.Fatalf("seed chat 200: %v", err)
		}
		return svc
	}

	t.Run("nil chatID returns all incidents", func(t *testing.T) {
		svc := setup()

		incidents, err := svc.ListIncidents(ctx, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(incidents) != 3 {
			t.Fatalf("expected 3 incidents, got %d", len(incidents))
		}
	})

	t.Run("chatID filters to a single chat", func(t *testing.T) {
		svc := setup()

		incidents, err := svc.ListIncidents(ctx, ptrInt64(100))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(incidents) != 2 {
			t.Fatalf("expected 2 incidents for chat 100, got %d", len(incidents))
		}
		for _, inc := range incidents {
			if inc.ChatID != 100 {
				t.Fatalf("expected only chat 100 incidents, got chat %d", inc.ChatID)
			}
		}
	})

	t.Run("unknown chatID returns no incidents", func(t *testing.T) {
		svc := setup()

		incidents, err := svc.ListIncidents(ctx, ptrInt64(999))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(incidents) != 0 {
			t.Fatalf("expected no incidents, got %d", len(incidents))
		}
	})
}
