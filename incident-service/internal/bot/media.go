package bot

import (
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"time"

	"gopkg.in/telebot.v3"

	"github.com/cQu1x/Incident-War-Room/internal/domain/media"
	"github.com/cQu1x/Incident-War-Room/internal/errs"
)

// maxMediaBytes caps how much of a single attachment is read into memory.
const maxMediaBytes = 20 << 20

// albumWindow is how long album items are buffered before the album is flushed
// as one timeline event. Telegram delivers the photos/videos of an album as
// separate messages that share an AlbumID, arriving back-to-back.
const albumWindow = 1500 * time.Millisecond

const mediaNotConnected = "⚠️ Media is not supported: the S3 storage is not connected."

const mediaUploadFailed = "⚠️ Couldn't attach the media. Please try again in a moment."

// album buffers the attachments of a single Telegram album until it is flushed
// as one timeline event.
type album struct {
	chatID   int64
	topicID  int64
	userID   *int64
	username string
	caption  string
	files    []media.File
	timer    *time.Timer
}

// HandleTopicMedia records media (photos, video, documents, voice, …) posted in
// an incident topic on the timeline. Any media type is accepted, and any number
// of attachments: single messages are recorded immediately, while albums are
// buffered briefly and recorded together as one event. When media uploads are
// disabled the sender is told that media is unsupported. Topics without an
// active incident, and media outside a topic, are ignored.
func (h *Handler) HandleTopicMedia(c telebot.Context) error {
	topicID := threadID(c)
	if topicID == 0 {
		return nil
	}

	ctx, cancel := reqContext()
	defer cancel()

	if _, _, err := h.svc.GetTimeline(ctx, c.Chat().ID, topicID); err != nil {
		if !errors.Is(err, errs.ErrNoActiveIncident) {
			log.Printf("bot: handle topic media: %v", err)
		}
		return nil
	}

	opts := &telebot.SendOptions{ThreadID: int(topicID)}

	if !h.mediaEnabled {
		return c.Send(mediaNotConnected, opts)
	}

	m := c.Message()
	file, err := h.downloadMedia(m)
	if err != nil {
		log.Printf("bot: download media: %v", err)
		return c.Send(mediaUploadFailed, opts)
	}

	userID, username := sender(c)

	// Album items share an AlbumID and are delivered as separate messages; buffer
	// them so the whole album lands on the timeline as a single event.
	if m.AlbumID != "" {
		h.bufferAlbumItem(c.Chat().ID, topicID, userID, username, m.Caption, m.AlbumID, file)
		return nil
	}

	if _, err := h.svc.AddTimelineEventWithMedia(ctx, c.Chat().ID, topicID, userID, username, m.Caption, []media.File{file}); err != nil {
		if errors.Is(err, errs.ErrNoActiveIncident) {
			return nil
		}
		log.Printf("bot: add timeline media: %v", err)
		return c.Send(mediaUploadFailed, opts)
	}

	h.refreshTimeline(c.Chat().ID, topicID)
	return nil
}

// bufferAlbumItem appends one attachment to the in-flight album keyed by
// albumID and (re)arms the flush timer. The first non-empty caption seen for the
// album is kept.
func (h *Handler) bufferAlbumItem(chatID, topicID int64, userID *int64, username, caption, albumID string, file media.File) {
	h.albumMu.Lock()
	defer h.albumMu.Unlock()

	a, ok := h.albums[albumID]
	if !ok {
		a = &album{chatID: chatID, topicID: topicID, userID: userID, username: username}
		h.albums[albumID] = a
	}
	a.files = append(a.files, file)
	if a.caption == "" && caption != "" {
		a.caption = caption
	}

	if a.timer != nil {
		a.timer.Stop()
	}
	a.timer = time.AfterFunc(h.albumWindow, func() { h.flushAlbum(albumID) })
}

// flushAlbum records the buffered album as one timeline event and drops it from
// the buffer.
func (h *Handler) flushAlbum(albumID string) {
	h.albumMu.Lock()
	a, ok := h.albums[albumID]
	if ok {
		delete(h.albums, albumID)
	}
	h.albumMu.Unlock()
	if !ok || len(a.files) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), handlerTimeout)
	defer cancel()

	if _, err := h.svc.AddTimelineEventWithMedia(ctx, a.chatID, a.topicID, a.userID, a.username, a.caption, a.files); err != nil {
		if !errors.Is(err, errs.ErrNoActiveIncident) {
			log.Printf("bot: add timeline album: %v", err)
		}
		return
	}

	h.refreshTimeline(a.chatID, a.topicID)
}

// downloadMedia fetches the message's attachment (whatever its type) into
// memory.
func (h *Handler) downloadMedia(m *telebot.Message) (media.File, error) {
	med := m.Media()
	if med == nil {
		return media.File{}, errors.New("message carries no media")
	}

	file, err := h.api.FileByID(med.MediaFile().FileID)
	if err != nil {
		return media.File{}, err
	}

	rc, err := h.api.File(&file)
	if err != nil {
		return media.File{}, err
	}
	defer rc.Close()

	data, err := io.ReadAll(io.LimitReader(rc, maxMediaBytes))
	if err != nil {
		return media.File{}, err
	}

	contentType := http.DetectContentType(data)
	return media.File{
		Data:        data,
		ContentType: contentType,
		Ext:         extForContentType(contentType),
	}, nil
}

func extForContentType(contentType string) string {
	switch contentType {
	case "image/png":
		return "png"
	case "image/gif":
		return "gif"
	case "image/webp":
		return "webp"
	case "image/jpeg":
		return "jpg"
	case "video/mp4":
		return "mp4"
	case "video/webm":
		return "webm"
	case "audio/mpeg":
		return "mp3"
	case "audio/ogg":
		return "ogg"
	case "application/pdf":
		return "pdf"
	default:
		return "bin"
	}
}
