package bot

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/google/uuid"
	"gopkg.in/telebot.v3"

	"github.com/cQu1x/Incident-War-Room/internal/domain/event"
	"github.com/cQu1x/Incident-War-Room/internal/domain/incident"
	"github.com/cQu1x/Incident-War-Room/internal/domain/media"
	"github.com/cQu1x/Incident-War-Room/internal/domain/report"
)

type mockContext struct {
	telebot.Context
	args        []string
	chatID      int64
	threadID    int64
	data        string
	user        *telebot.User
	message     *telebot.Message
	sent        []string
	sentThread  []int
	editedReply []*telebot.ReplyMarkup
}

func (m *mockContext) Args() []string { return m.args }

func (m *mockContext) Data() string { return m.data }

func (m *mockContext) Chat() *telebot.Chat { return &telebot.Chat{ID: m.chatID} }

func (m *mockContext) Message() *telebot.Message {
	if m.message != nil {
		return m.message
	}
	return &telebot.Message{ThreadID: int(m.threadID)}
}

func (m *mockContext) Sender() *telebot.User { return m.user }

func (m *mockContext) Respond(_ ...*telebot.CallbackResponse) error { return nil }

func (m *mockContext) Edit(what interface{}, _ ...interface{}) error {
	if markup, ok := what.(*telebot.ReplyMarkup); ok {
		m.editedReply = append(m.editedReply, markup)
	}
	return nil
}

func (m *mockContext) Send(what interface{}, opts ...interface{}) error {
	if s, ok := what.(string); ok {
		m.sent = append(m.sent, s)
	} else {
		m.sent = append(m.sent, fmt.Sprintf("<%T>", what))
	}

	var thread int
	for _, o := range opts {
		if so, ok := o.(*telebot.SendOptions); ok {
			thread = so.ThreadID
		}
	}
	m.sentThread = append(m.sentThread, thread)
	return nil
}

func lastSent(t *testing.T, m *mockContext) string {
	t.Helper()
	if len(m.sent) != 1 {
		t.Fatalf("expected exactly 1 message sent, got %d: %v", len(m.sent), m.sent)
	}
	return m.sent[0]
}

func sentContains(t *testing.T, m *mockContext, substr string) {
	t.Helper()
	for _, s := range m.sent {
		if strings.Contains(s, substr) {
			return
		}
	}
	t.Fatalf("no sent message contains %q; sent: %v", substr, m.sent)
}

type fakeService struct {
	create   func(chatID, topicID int64, title string, sev incident.Severity, userID *int64, username string) (*incident.Incident, error)
	addEvent func(chatID, topicID int64, userID *int64, username, message string) (*event.Event, error)
	addImage func(chatID, topicID int64, userID *int64, username, caption string, img media.Image) (*event.Event, error)
	closeInc func(chatID, topicID int64, userID *int64, username string) (*incident.Incident, error)
	reopen   func(id uuid.UUID, newTopicID int64, userID *int64, username string) (*incident.Incident, error)
	getInc   func(id uuid.UUID) (*incident.Incident, error)
	setSev   func(chatID, topicID int64, sev incident.Severity) (*incident.Incident, error)
	timeline func(chatID, topicID int64) (*incident.Incident, []event.Event, error)
	publish  func(chatID, topicID int64) ([]string, error)
	report   func(chatID, topicID int64) (report.Document, error)
}

func (f *fakeService) CreateIncident(_ context.Context, chatID, topicID int64, title string, sev incident.Severity, userID *int64, username string) (*incident.Incident, error) {
	return f.create(chatID, topicID, title, sev, userID, username)
}

func (f *fakeService) AddTimelineEvent(_ context.Context, chatID, topicID int64, userID *int64, username, message string) (*event.Event, error) {
	return f.addEvent(chatID, topicID, userID, username, message)
}

func (f *fakeService) AddTimelineEventWithImage(_ context.Context, chatID, topicID int64, userID *int64, username, caption string, img media.Image) (*event.Event, error) {
	return f.addImage(chatID, topicID, userID, username, caption, img)
}

func (f *fakeService) CloseIncident(_ context.Context, chatID, topicID int64, userID *int64, username string) (*incident.Incident, error) {
	return f.closeInc(chatID, topicID, userID, username)
}

func (f *fakeService) ReopenIncident(_ context.Context, id uuid.UUID, newTopicID int64, userID *int64, username string) (*incident.Incident, error) {
	return f.reopen(id, newTopicID, userID, username)
}

func (f *fakeService) GetIncident(_ context.Context, id uuid.UUID) (*incident.Incident, error) {
	return f.getInc(id)
}

func (f *fakeService) SetSeverity(_ context.Context, chatID, topicID int64, sev incident.Severity) (*incident.Incident, error) {
	return f.setSev(chatID, topicID, sev)
}

func (f *fakeService) GetTimeline(_ context.Context, chatID, topicID int64) (*incident.Incident, []event.Event, error) {
	return f.timeline(chatID, topicID)
}

func (f *fakeService) PublishTimeline(_ context.Context, chatID, topicID int64) ([]string, error) {
	if f.publish == nil {
		return nil, nil
	}
	return f.publish(chatID, topicID)
}

func (f *fakeService) GenerateReport(_ context.Context, chatID, topicID int64) (report.Document, error) {
	return f.report(chatID, topicID)
}

type fakeAPI struct {
	createdTopic *telebot.Topic
	createErr    error
	deleted      []int
	sent         []sentMessage
	edited       []sentMessage
}

type sentMessage struct {
	threadID int
	what     string
	markup   *telebot.ReplyMarkup
}

func newFakeAPI() *fakeAPI {
	return &fakeAPI{createdTopic: &telebot.Topic{Name: "topic", ThreadID: 555}}
}

func (a *fakeAPI) Send(_ telebot.Recipient, what interface{}, opts ...interface{}) (*telebot.Message, error) {
	var thread int
	var markup *telebot.ReplyMarkup
	for _, o := range opts {
		switch v := o.(type) {
		case *telebot.SendOptions:
			thread = v.ThreadID
			markup = v.ReplyMarkup
		case *telebot.ReplyMarkup:
			markup = v
		}
	}

	msg := sentMessage{threadID: thread, markup: markup}
	if s, ok := what.(string); ok {
		msg.what = s
	} else {
		msg.what = fmt.Sprintf("<%T>", what)
	}
	a.sent = append(a.sent, msg)
	return &telebot.Message{}, nil
}

func (a *fakeAPI) Edit(_ telebot.Editable, what interface{}, opts ...interface{}) (*telebot.Message, error) {
	var markup *telebot.ReplyMarkup
	for _, o := range opts {
		if v, ok := o.(*telebot.ReplyMarkup); ok {
			markup = v
		}
	}

	msg := sentMessage{markup: markup}
	if s, ok := what.(string); ok {
		msg.what = s
	} else {
		msg.what = fmt.Sprintf("<%T>", what)
	}
	a.edited = append(a.edited, msg)
	return &telebot.Message{}, nil
}

func (a *fakeAPI) CreateTopic(_ *telebot.Chat, topic *telebot.Topic) (*telebot.Topic, error) {
	if a.createErr != nil {
		return nil, a.createErr
	}
	a.createdTopic.Name = topic.Name
	return a.createdTopic, nil
}

func (a *fakeAPI) DeleteTopic(_ *telebot.Chat, topic *telebot.Topic) error {
	a.deleted = append(a.deleted, topic.ThreadID)
	return nil
}

func (a *fakeAPI) FileByID(fileID string) (telebot.File, error) {
	return telebot.File{FileID: fileID, FilePath: "photos/" + fileID + ".jpg"}, nil
}

func (a *fakeAPI) File(_ *telebot.File) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("\xff\xd8\xff\xe0image-bytes")), nil
}

func apiSentContains(t *testing.T, a *fakeAPI, substr string) {
	t.Helper()
	for _, s := range a.sent {
		if strings.Contains(s.what, substr) {
			return
		}
	}
	t.Fatalf("no api-sent message contains %q; sent: %v", substr, a.sent)
}
