package bot

import (
	"sync"
	"testing"
	"time"

	"gopkg.in/telebot.v3"

	"github.com/cQu1x/Incident-War-Room/internal/domain/event"
	"github.com/cQu1x/Incident-War-Room/internal/domain/incident"
	"github.com/cQu1x/Incident-War-Room/internal/domain/media"
	"github.com/cQu1x/Incident-War-Room/internal/errs"
)

func activeTimeline() func(int64, int64) (*incident.Incident, []event.Event, error) {
	return func(int64, int64) (*incident.Incident, []event.Event, error) {
		return &incident.Incident{}, nil, nil
	}
}

func TestHandleTopicMediaDisabled(t *testing.T) {
	called := false
	h := New(&fakeService{
		timeline: activeTimeline(),
		addMedia: func(int64, int64, *int64, string, string, []media.File) (*event.Event, error) {
			called = true
			return &event.Event{}, nil
		},
	}, newFakeAPI())

	ctx := &mockContext{chatID: 42, message: &telebot.Message{ThreadID: 7, Photo: &telebot.Photo{}}}

	if err := h.HandleTopicMedia(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sentContains(t, ctx, "S3 storage is not connected")
	if called {
		t.Error("media should not be uploaded when media is disabled")
	}
}

func TestHandleTopicMediaUploadsSingle(t *testing.T) {
	var gotCaption string
	var gotFiles []media.File
	h := New(&fakeService{
		timeline: activeTimeline(),
		addMedia: func(_, _ int64, _ *int64, _, caption string, files []media.File) (*event.Event, error) {
			gotCaption = caption
			gotFiles = files
			return &event.Event{}, nil
		},
	}, newFakeAPI(), WithMediaEnabled(true))

	ctx := &mockContext{chatID: 42, message: &telebot.Message{
		ThreadID: 7,
		Caption:  "prod down",
		Photo:    &telebot.Photo{File: telebot.File{FileID: "abc"}},
	}}

	if err := h.HandleTopicMedia(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotCaption != "prod down" {
		t.Errorf("caption = %q, want %q", gotCaption, "prod down")
	}
	if len(gotFiles) != 1 || len(gotFiles[0].Data) == 0 || gotFiles[0].ContentType != "image/jpeg" || gotFiles[0].Ext != "jpg" {
		t.Errorf("unexpected files: %+v", gotFiles)
	}
	if len(ctx.sent) != 0 {
		t.Errorf("expected no reply on success, got %v", ctx.sent)
	}
}

func TestHandleTopicMediaAcceptsNonPhoto(t *testing.T) {
	var gotFiles []media.File
	h := New(&fakeService{
		timeline: activeTimeline(),
		addMedia: func(_, _ int64, _ *int64, _, _ string, files []media.File) (*event.Event, error) {
			gotFiles = files
			return &event.Event{}, nil
		},
	}, newFakeAPI(), WithMediaEnabled(true))

	ctx := &mockContext{chatID: 42, message: &telebot.Message{
		ThreadID: 7,
		Video:    &telebot.Video{File: telebot.File{FileID: "vid"}},
	}}

	if err := h.HandleTopicMedia(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(gotFiles) != 1 {
		t.Fatalf("expected the video to be recorded, got %d files", len(gotFiles))
	}
}

func TestHandleTopicMediaBuffersAlbum(t *testing.T) {
	var mu sync.Mutex
	var gotFiles []media.File
	done := make(chan struct{})
	h := New(&fakeService{
		timeline: activeTimeline(),
		addMedia: func(_, _ int64, _ *int64, _, _ string, files []media.File) (*event.Event, error) {
			mu.Lock()
			gotFiles = files
			mu.Unlock()
			close(done)
			return &event.Event{}, nil
		},
	}, newFakeAPI(), WithMediaEnabled(true), WithAlbumWindow(20*time.Millisecond))

	for i := 0; i < 3; i++ {
		ctx := &mockContext{chatID: 42, message: &telebot.Message{
			ThreadID: 7,
			AlbumID:  "grp1",
			Caption:  "outage shots",
			Photo:    &telebot.Photo{File: telebot.File{FileID: "abc"}},
		}}
		if err := h.HandleTopicMedia(ctx); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(ctx.sent) != 0 {
			t.Errorf("album items should not reply, got %v", ctx.sent)
		}
	}

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("album was never flushed")
	}

	mu.Lock()
	defer mu.Unlock()
	if len(gotFiles) != 3 {
		t.Fatalf("expected 3 files in the album event, got %d", len(gotFiles))
	}
}

func TestHandleTopicMediaNoActiveIncidentSilent(t *testing.T) {
	h := New(&fakeService{
		timeline: func(int64, int64) (*incident.Incident, []event.Event, error) {
			return nil, nil, errs.ErrNoActiveIncident
		},
	}, newFakeAPI(), WithMediaEnabled(true))

	ctx := &mockContext{chatID: 42, message: &telebot.Message{ThreadID: 7, Photo: &telebot.Photo{}}}

	if err := h.HandleTopicMedia(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ctx.sent) != 0 {
		t.Errorf("expected no reply without an active incident, got %v", ctx.sent)
	}
}
