package bot

import (
	"testing"
	"time"

	"github.com/cQu1x/Incident-War-Room/internal/domain/event"
	"github.com/cQu1x/Incident-War-Room/internal/domain/incident"
)

func waitTopic(t *testing.T, calls <-chan int64, want int64) {
	t.Helper()
	select {
	case got := <-calls:
		if got != want {
			t.Fatalf("refreshed topic %d, want %d", got, want)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for the async timeline refresh")
	}
}

func TestRefreshTimelinePublishesAsync(t *testing.T) {
	calls := make(chan int64, 1)
	h := New(&fakeService{
		publish: func(_, topicID int64) ([]string, error) {
			calls <- topicID
			return []string{"https://telegra.ph/p1"}, nil
		},
	}, newFakeAPI())

	h.refreshTimeline(100, 7)
	waitTopic(t, calls, 7)
}

func TestRefreshTimelineIgnoresGeneralChat(t *testing.T) {
	h := New(&fakeService{
		publish: func(int64, int64) ([]string, error) {
			t.Fatal("timeline refresh should be skipped outside a topic")
			return nil, nil
		},
	}, newFakeAPI())

	h.refreshTimeline(100, 0)
	time.Sleep(50 * time.Millisecond)
}

func TestAddUpdateRefreshesTimeline(t *testing.T) {
	calls := make(chan int64, 1)
	h := New(&fakeService{
		addEvent: func(int64, int64, *int64, string, string) (*event.Event, error) {
			return &event.Event{}, nil
		},
		publish: func(_, topicID int64) ([]string, error) {
			calls <- topicID
			return nil, nil
		},
	}, newFakeAPI())

	if err := h.HandleIncident(&mockContext{args: []string{"db", "down"}, chatID: 100, threadID: 9}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	waitTopic(t, calls, 9)
}

func TestSetSeverityRefreshesTimeline(t *testing.T) {
	calls := make(chan int64, 1)
	h := New(&fakeService{
		setSev: func(_, topicID int64, sev incident.Severity, _ *int64, _ string) (*incident.Incident, error) {
			return &incident.Incident{Title: "outage", TopicID: topicID, Severity: sev, Status: incident.StatusActive}, nil
		},
		publish: func(_, topicID int64) ([]string, error) {
			calls <- topicID
			return nil, nil
		},
	}, newFakeAPI())

	ctx := &mockContext{chatID: 100, threadID: 5, data: string(incident.SeverityHigh)}
	if err := h.handleSetSeverity(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	waitTopic(t, calls, 5)
}
