package bot

import (
	"testing"

	"gopkg.in/telebot.v3"

	"github.com/cQu1x/Incident-War-Room/internal/domain/event"
	"github.com/cQu1x/Incident-War-Room/internal/domain/incident"
	"github.com/cQu1x/Incident-War-Room/internal/errs"
)

func TestHandleTopicMediaOutsideTopicIgnored(t *testing.T) {
	called := false
	h := New(&fakeService{
		timeline: func(int64, int64) (*incident.Incident, []event.Event, error) {
			called = true
			return &incident.Incident{}, nil, nil
		},
	}, newFakeAPI())

	ctx := &mockContext{message: &telebot.Message{ThreadID: 0, Photo: &telebot.Photo{}}}

	if err := h.HandleTopicMedia(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Error("GetTimeline should not be called outside a topic")
	}
	if len(ctx.sent) != 0 {
		t.Errorf("expected no reply outside a topic, got %v", ctx.sent)
	}
}

func TestHandleTopicTextRecordsMessage(t *testing.T) {
	var gotChat, gotTopic int64
	var gotMessage string
	h := New(&fakeService{
		addEvent: func(chatID, topicID int64, _ *int64, _, message string) (*event.Event, error) {
			gotChat, gotTopic, gotMessage = chatID, topicID, message
			return &event.Event{}, nil
		},
	}, newFakeAPI())

	ctx := &mockContext{
		chatID:  42,
		message: &telebot.Message{ThreadID: 7, Text: "db is on fire"},
	}

	if err := h.HandleTopicText(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotChat != 42 || gotTopic != 7 {
		t.Errorf("recorded chat/topic = %d/%d, want 42/7", gotChat, gotTopic)
	}
	if gotMessage != "db is on fire" {
		t.Errorf("recorded message = %q, want %q", gotMessage, "db is on fire")
	}
	if len(ctx.sent) != 0 {
		t.Errorf("expected no reply, got %v", ctx.sent)
	}
}

func TestHandleTopicTextOutsideTopicIgnored(t *testing.T) {
	called := false
	h := New(&fakeService{
		addEvent: func(int64, int64, *int64, string, string) (*event.Event, error) {
			called = true
			return &event.Event{}, nil
		},
	}, newFakeAPI())

	ctx := &mockContext{message: &telebot.Message{ThreadID: 0, Text: "hello"}}

	if err := h.HandleTopicText(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Error("AddTimelineEvent should not be called outside a topic")
	}
}

func TestHandleTopicTextNoActiveIncidentSilent(t *testing.T) {
	h := New(&fakeService{
		addEvent: func(int64, int64, *int64, string, string) (*event.Event, error) {
			return nil, errs.ErrNoActiveIncident
		},
	}, newFakeAPI())

	ctx := &mockContext{message: &telebot.Message{ThreadID: 7, Text: "hello"}}

	if err := h.HandleTopicText(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ctx.sent) != 0 {
		t.Errorf("expected no reply, got %v", ctx.sent)
	}
}

func TestHandleTopicTextIgnoresMediaCaption(t *testing.T) {
	for name, msg := range map[string]*telebot.Message{
		"photo":    {ThreadID: 7, Photo: &telebot.Photo{}, Caption: "screenshot"},
		"voice":    {ThreadID: 7, Voice: &telebot.Voice{}, Caption: "voice note"},
		"document": {ThreadID: 7, Document: &telebot.Document{}, Caption: "logs attached"},
	} {
		t.Run(name, func(t *testing.T) {
			called := false
			h := New(&fakeService{
				addEvent: func(int64, int64, *int64, string, string) (*event.Event, error) {
					called = true
					return &event.Event{}, nil
				},
			}, newFakeAPI())

			ctx := &mockContext{message: msg}

			if err := h.HandleTopicText(ctx); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if called {
				t.Errorf("%s caption should not be recorded on the timeline", name)
			}
		})
	}
}
