package bot

import (
	"errors"
	"io"
	"log"
	"net/http"

	"gopkg.in/telebot.v3"

	"github.com/cQu1x/Incident-War-Room/internal/domain/media"
	"github.com/cQu1x/Incident-War-Room/internal/errs"
)

// maxImageBytes caps how much of a photo download is read into memory.
const maxImageBytes = 20 << 20

const imagesNotConnected = "⚠️ Images are not supported: the S3 storage is not connected."

const tooManyPhotos = "⚠️ Only one photo per message is allowed. Please send a single image."

const imageUploadFailed = "⚠️ Couldn't attach the image. Please try again in a moment."

// HandleTopicPhoto records a single photo posted in an incident topic on the
// timeline. When media uploads are disabled the sender is told that images are
// unsupported. Albums (more than one photo) are rejected; only one photo per
// message is allowed. Topics without an active incident, and photos outside a
// topic, are ignored.
func (h *Handler) HandleTopicPhoto(c telebot.Context) error {
	topicID := threadID(c)
	if topicID == 0 {
		return nil
	}

	ctx, cancel := reqContext()
	defer cancel()

	if _, _, err := h.svc.GetTimeline(ctx, c.Chat().ID, topicID); err != nil {
		if !errors.Is(err, errs.ErrNoActiveIncident) {
			log.Printf("bot: handle topic photo: %v", err)
		}
		return nil
	}

	opts := &telebot.SendOptions{ThreadID: int(topicID)}

	if !h.mediaEnabled {
		return c.Send(imagesNotConnected, opts)
	}

	m := c.Message()
	if m.AlbumID != "" {
		return c.Send(tooManyPhotos, opts)
	}

	img, err := h.downloadPhoto(m.Photo)
	if err != nil {
		log.Printf("bot: download photo: %v", err)
		return c.Send(imageUploadFailed, opts)
	}

	userID, username := sender(c)
	if _, err := h.svc.AddTimelineEventWithImage(ctx, c.Chat().ID, topicID, userID, username, m.Caption, img); err != nil {
		if errors.Is(err, errs.ErrNoActiveIncident) {
			return nil
		}
		log.Printf("bot: add timeline image: %v", err)
		return c.Send(imageUploadFailed, opts)
	}

	h.refreshTimeline(c.Chat().ID, topicID)
	return nil
}

func (h *Handler) downloadPhoto(photo *telebot.Photo) (media.Image, error) {
	file, err := h.api.FileByID(photo.FileID)
	if err != nil {
		return media.Image{}, err
	}

	rc, err := h.api.File(&file)
	if err != nil {
		return media.Image{}, err
	}
	defer rc.Close()

	data, err := io.ReadAll(io.LimitReader(rc, maxImageBytes))
	if err != nil {
		return media.Image{}, err
	}

	contentType := http.DetectContentType(data)
	return media.Image{
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
	default:
		return "jpg"
	}
}
