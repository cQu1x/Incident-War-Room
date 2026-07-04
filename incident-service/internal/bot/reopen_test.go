package bot

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/cQu1x/Incident-War-Room/internal/domain/incident"
	"github.com/cQu1x/Incident-War-Room/internal/errs"
)

func TestHandleReopenOpensNewTopicAndAnnounces(t *testing.T) {
	id := uuid.New()
	api := newFakeAPI()
	api.createdTopic.ThreadID = 909

	var gotID uuid.UUID
	var gotTopic int64
	h := New(&fakeService{
		getInc: func(uuid.UUID) (*incident.Incident, error) {
			return &incident.Incident{ID: id, Title: "db down", ChatID: 100, Status: incident.StatusClosed}, nil
		},
		reopen: func(reopenID uuid.UUID, newTopicID int64, _ *int64, _ string) (*incident.Incident, error) {
			gotID, gotTopic = reopenID, newTopicID
			return &incident.Incident{ID: id, Title: "db down", ChatID: 100, TopicID: newTopicID, Severity: incident.SeverityHigh, Status: incident.StatusActive}, nil
		},
	}, api)
	ctx := &mockContext{chatID: 100, data: id.String()}

	if err := h.handleReopenIncident(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotID != id {
		t.Errorf("service reopened %s, want %s", gotID, id)
	}
	if gotTopic != 909 {
		t.Errorf("service got topic id %d, want 909", gotTopic)
	}
	if len(api.sent) != 2 {
		t.Fatalf("expected the topic card and the announcement, got %v", api.sent)
	}

	card := api.sent[0]
	if card.threadID != 909 {
		t.Errorf("card sent to thread %d, want the new topic 909", card.threadID)
	}
	if m := card.markup; m == nil || len(m.InlineKeyboard) == 0 {
		t.Errorf("card sent without an inline menu: %+v", m)
	}

	announce := api.sent[1]
	if announce.threadID != 0 {
		t.Errorf("announcement sent to thread %d, want main chat (0)", announce.threadID)
	}
	if !strings.Contains(announce.what, "Incident reopened") {
		t.Errorf("announcement %q is not the reopen summary", announce.what)
	}

	if len(ctx.editedReply) != 1 || ctx.editedReply[0] == nil || len(ctx.editedReply[0].InlineKeyboard) != 0 {
		t.Errorf("expected the reopen button to be cleared, got %+v", ctx.editedReply)
	}
}

func TestHandleReopenRejectsForeignChat(t *testing.T) {
	id := uuid.New()
	api := newFakeAPI()
	h := New(&fakeService{
		getInc: func(uuid.UUID) (*incident.Incident, error) {
			return &incident.Incident{ID: id, Title: "db down", ChatID: 999, Status: incident.StatusClosed}, nil
		},
		reopen: func(uuid.UUID, int64, *int64, string) (*incident.Incident, error) {
			t.Fatal("reopen should not run for a foreign chat")
			return nil, nil
		},
	}, api)
	ctx := &mockContext{chatID: 100, data: id.String()}

	if err := h.handleReopenIncident(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(api.sent) != 0 {
		t.Errorf("expected no topic to be created, got %v", api.sent)
	}
}

func TestHandleReopenDeletesOrphanTopicOnConflict(t *testing.T) {
	id := uuid.New()
	api := newFakeAPI()
	api.createdTopic.ThreadID = 42
	h := New(&fakeService{
		getInc: func(uuid.UUID) (*incident.Incident, error) {
			return &incident.Incident{ID: id, Title: "db down", ChatID: 100, Status: incident.StatusClosed}, nil
		},
		reopen: func(uuid.UUID, int64, *int64, string) (*incident.Incident, error) {
			return nil, errs.ErrIncidentAlreadyActive
		},
	}, api)
	ctx := &mockContext{chatID: 100, data: id.String()}

	if err := h.handleReopenIncident(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sentContains(t, ctx, "already active")
	if len(api.deleted) != 1 || api.deleted[0] != 42 {
		t.Fatalf("expected orphan topic 42 to be deleted, got %v", api.deleted)
	}
	if len(ctx.editedReply) != 0 {
		t.Errorf("expected the reopen button to stay on failure, got %+v", ctx.editedReply)
	}
}

func TestHandleReopenRejectsInvalidData(t *testing.T) {
	h := New(&fakeService{
		getInc: func(uuid.UUID) (*incident.Incident, error) {
			t.Fatal("service should not be called for invalid data")
			return nil, nil
		},
	}, newFakeAPI())
	ctx := &mockContext{chatID: 100, data: "not-a-uuid"}

	if err := h.handleReopenIncident(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sentContains(t, ctx, "can no longer be reopened")
}
