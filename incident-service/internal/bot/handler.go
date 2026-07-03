package bot

import (
	"context"
	"errors"
	"io"
	"sync"
	"time"

	"github.com/google/uuid"
	"gopkg.in/telebot.v3"

	"github.com/cQu1x/Incident-War-Room/internal/domain/event"
	"github.com/cQu1x/Incident-War-Room/internal/domain/incident"
	"github.com/cQu1x/Incident-War-Room/internal/domain/media"
	"github.com/cQu1x/Incident-War-Room/internal/domain/report"
	"github.com/cQu1x/Incident-War-Room/internal/errs"
)

const handlerTimeout = 30 * time.Second

type IncidentService interface {
	CreateIncident(ctx context.Context, chatID, topicID int64, title string, severity incident.Severity, userID *int64, username string) (*incident.Incident, error)
	AddTimelineEvent(ctx context.Context, chatID, topicID int64, userID *int64, username, message string) (*event.Event, error)
	AddTimelineEventWithImage(ctx context.Context, chatID, topicID int64, userID *int64, username, caption string, img media.Image) (*event.Event, error)
	CloseIncident(ctx context.Context, chatID, topicID int64, userID *int64, username string) (*incident.Incident, error)
	SetSeverity(ctx context.Context, chatID, topicID int64, severity incident.Severity) (*incident.Incident, error)
	GetTimeline(ctx context.Context, chatID, topicID int64) (*incident.Incident, []event.Event, error)
	PublishTimeline(ctx context.Context, chatID, topicID int64) ([]string, error)
	GenerateReport(ctx context.Context, chatID, topicID int64) (report.Document, error)
}

// DashboardLinker builds the tokenised web-dashboard link an operator receives
// when an incident is closed.
type DashboardLinker interface {
	Link(incidentID uuid.UUID) (string, error)
}

type TelegramAPI interface {
	Send(to telebot.Recipient, what interface{}, opts ...interface{}) (*telebot.Message, error)
	Edit(msg telebot.Editable, what interface{}, opts ...interface{}) (*telebot.Message, error)
	CreateTopic(chat *telebot.Chat, topic *telebot.Topic) (*telebot.Topic, error)
	DeleteTopic(chat *telebot.Chat, topic *telebot.Topic) error
	FileByID(fileID string) (telebot.File, error)
	File(file *telebot.File) (io.ReadCloser, error)
}

type announceKey struct {
	chatID  int64
	topicID int64
}

type Handler struct {
	svc          IncidentService
	api          TelegramAPI
	dashboard    DashboardLinker
	mediaEnabled bool
	alertChatID  int64

	mu            sync.Mutex
	announcements map[announceKey]telebot.Editable
}

// Option customizes a Handler.
type Option func(*Handler)

// WithMediaEnabled toggles image uploads in incident topics. When disabled the
// bot replies that images are unsupported because S3 storage is not connected.
func WithMediaEnabled(enabled bool) Option {
	return func(h *Handler) { h.mediaEnabled = enabled }
}

// WithDashboard wires the linker that mints the tokenised dashboard link
// included in the incident-closed message. When unset, no link is shown.
func WithDashboard(linker DashboardLinker) Option {
	return func(h *Handler) { h.dashboard = linker }
}

// WithAlertChat sets the forum supergroup where incidents opened from external
// monitoring alerts are created. When unset, OpenIncidentFromAlert is rejected.
func WithAlertChat(chatID int64) Option {
	return func(h *Handler) { h.alertChatID = chatID }
}

func New(svc IncidentService, api TelegramAPI, opts ...Option) *Handler {
	h := &Handler{
		svc:           svc,
		api:           api,
		announcements: make(map[announceKey]telebot.Editable),
	}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

func (h *Handler) rememberAnnouncement(chatID, topicID int64, msg telebot.Editable) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.announcements[announceKey{chatID, topicID}] = msg
}

func (h *Handler) announcement(chatID, topicID int64) (telebot.Editable, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	msg, ok := h.announcements[announceKey{chatID, topicID}]
	return msg, ok
}

func (h *Handler) forgetAnnouncement(chatID, topicID int64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.announcements, announceKey{chatID, topicID})
}

func reqContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), handlerTimeout)
}

func threadID(c telebot.Context) int64 {
	if m := c.Message(); m != nil {
		return int64(m.ThreadID)
	}
	return 0
}

func sender(c telebot.Context) (*int64, string) {
	u := c.Sender()
	if u == nil {
		return nil, ""
	}
	id := u.ID
	return &id, u.Username
}

func userError(err error) string {
	switch {
	case errors.Is(err, errs.ErrNoActiveIncident):
		return "There is no active incident in this chat. Open one with /incident create <description>."
	case errors.Is(err, errs.ErrIncidentAlreadyActive):
		return "An incident is already active in this chat. Close it before opening a new one."
	case errs.Is(err, errs.KindValidation):
		return "Sorry, that input is not valid. " + incidentUsage
	case errs.Is(err, errs.KindUnavailable):
		return "The service is temporarily unavailable. Please try again in a moment."
	default:
		return "Something went wrong. Please try again."
	}
}
