package bot

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"gopkg.in/telebot.v3"

	"github.com/cQu1x/Incident-War-Room/internal/bot/response"
	"github.com/cQu1x/Incident-War-Room/internal/domain/incident"
	"github.com/cQu1x/Incident-War-Room/internal/domain/report"
	"github.com/cQu1x/Incident-War-Room/internal/errs"
)

var errCreateTopic = errors.New("create topic")

const incidentUsage = "Usage:\n/incident create <description> — open a new incident\n/incident close — close the active incident\n/incident <message> — add an update to the timeline"

const topicNameLimit = 128

const topicForumRequired = "Couldn't open a topic for this incident. Use /incident create in a forum supergroup where the bot is an admin with the \"Manage Topics\" right."

func topicName(title string) string {
	r := []rune(title)
	if len(r) > topicNameLimit {
		return string(r[:topicNameLimit])
	}
	return title
}

func topicLink(chat *telebot.Chat, threadID int) string {
	if chat.Username != "" {
		return fmt.Sprintf("https://t.me/%s/%d", chat.Username, threadID)
	}
	id := strings.TrimPrefix(strconv.FormatInt(chat.ID, 10), "-100")
	return fmt.Sprintf("https://t.me/c/%s/%d", id, threadID)
}

func (h *Handler) HandleIncident(c telebot.Context) error {
	args := c.Args()
	if len(args) == 0 {
		return c.Send(incidentUsage)
	}

	switch args[0] {
	case "create":
		return h.createIncident(c, strings.TrimSpace(strings.Join(args[1:], " ")))
	case "close":
		_, err := h.closeIncident(c)
		return err
	default:
		return h.addUpdate(c, strings.TrimSpace(strings.Join(args, " ")))
	}
}

func (h *Handler) createIncident(c telebot.Context, description string) error {
	if description == "" {
		return c.Send("Please add a description:\n/incident create <what happened>")
	}

	ctx, cancel := reqContext()
	defer cancel()

	userID, username := sender(c)
	_, err := h.openIncident(ctx, c.Chat(), description, "", userID, username)
	if errors.Is(err, errCreateTopic) {
		return c.Send(topicForumRequired)
	}
	if err != nil {
		return c.Send(userError(err))
	}
	return nil
}

// OpenIncidentFromAlert opens an incident in the configured alert chat from an
// external monitoring alert (e.g. an Alertmanager webhook). It returns
// errs.KindUnavailable when no alert chat is configured.
func (h *Handler) OpenIncidentFromAlert(ctx context.Context, title string, severity incident.Severity) (*incident.Incident, error) {
	if h.alertChatID == 0 {
		return nil, errs.New(errs.KindUnavailable, "bot.OpenIncidentFromAlert", "alert chat is not configured")
	}
	return h.openIncident(ctx, &telebot.Chat{ID: h.alertChatID}, title, severity, nil, "alertmanager")
}

func (h *Handler) openIncident(ctx context.Context, chat *telebot.Chat, title string, severity incident.Severity, userID *int64, username string) (*incident.Incident, error) {
	topic, err := h.api.CreateTopic(chat, &telebot.Topic{Name: topicName(title)})
	if err != nil {
		log.Printf("bot: create topic: %v", err)
		return nil, fmt.Errorf("%w: %v", errCreateTopic, err)
	}

	inc, err := h.svc.CreateIncident(ctx, chat.ID, int64(topic.ThreadID), title, severity, userID, username)
	if err != nil {
		log.Printf("bot: create incident: %v", err)
		if delErr := h.api.DeleteTopic(chat, topic); delErr != nil {
			log.Printf("bot: delete orphan topic %d: %v", topic.ThreadID, delErr)
		}
		return nil, err
	}

	if _, err := h.api.Send(
		chat,
		incidentCard(inc.Title, inc.Severity, inc.Status, h.mediaEnabled),
		&telebot.SendOptions{ThreadID: topic.ThreadID, ReplyMarkup: incidentMenu()},
	); err != nil {
		return nil, err
	}

	announcement, err := h.api.Send(
		chat,
		response.IncidentCreated(*inc, topicLink(chat, topic.ThreadID)),
		telebot.ModeHTML,
	)
	if err != nil {
		return nil, err
	}

	h.rememberAnnouncement(chat.ID, int64(topic.ThreadID), announcement)
	return inc, nil
}

func (h *Handler) refreshAnnouncement(c telebot.Context, inc incident.Incident) {
	chat := c.Chat()
	topicID := threadID(c)

	msg, ok := h.announcement(chat.ID, topicID)
	if !ok {
		return
	}

	if _, err := h.api.Edit(
		msg,
		response.IncidentCreated(inc, topicLink(chat, int(topicID))),
		telebot.ModeHTML,
	); err != nil {
		log.Printf("bot: refresh main-chat announcement: %v", err)
	}
}

func (h *Handler) addUpdate(c telebot.Context, message string) error {
	ctx, cancel := reqContext()
	defer cancel()

	userID, username := sender(c)
	if _, err := h.svc.AddTimelineEvent(ctx, c.Chat().ID, threadID(c), userID, username, message); err != nil {
		log.Printf("bot: add timeline event: %v", err)
		return c.Send(userError(err))
	}

	return c.Send("📝 Update added to the timeline.")
}

func (h *Handler) closeIncident(c telebot.Context) (*incident.Incident, error) {
	ctx, cancel := reqContext()
	defer cancel()

	chat := c.Chat()
	topicID := threadID(c)
	userID, username := sender(c)

	doc, reportErr := h.svc.GenerateReport(ctx, chat.ID, topicID)
	timelineURLs, pubErr := h.svc.PublishTimeline(ctx, chat.ID, topicID)

	inc, err := h.svc.CloseIncident(ctx, chat.ID, topicID, userID, username)
	if err != nil {
		log.Printf("bot: close incident: %v", err)
		return nil, c.Send(userError(err))
	}

	if pubErr != nil {
		log.Printf("bot: publish timeline: %v", pubErr)
		timelineURLs = nil
	}

	if reportErr != nil {
		log.Printf("bot: generate report: %v", reportErr)
		doc = report.Document{}
	}

	dashboardURL := h.dashboardLink(*inc)

	if _, err := h.api.Send(chat, response.IncidentClosed(*inc, timelineURLs, doc, dashboardURL), telebot.ModeHTML); err != nil {
		return inc, err
	}

	if len(doc.PDF) > 0 {
		if err := h.sendReportDocument(chat, *inc, doc.PDF); err != nil {
			log.Printf("bot: send report document: %v", err)
		}
	}

	if topicID != 0 {
		if err := h.api.DeleteTopic(chat, &telebot.Topic{ThreadID: int(topicID)}); err != nil {
			log.Printf("bot: delete topic %d: %v", topicID, err)
		}
	}

	h.forgetAnnouncement(chat.ID, topicID)
	return inc, nil
}

func (h *Handler) dashboardLink(inc incident.Incident) string {
	if h.dashboard == nil {
		return ""
	}
	link, err := h.dashboard.Link(inc.ID)
	if err != nil {
		log.Printf("bot: build dashboard link: %v", err)
		return ""
	}
	return link
}

func (h *Handler) sendReportDocument(chat *telebot.Chat, inc incident.Incident, pdf []byte) error {
	document := &telebot.Document{
		File:     telebot.FromReader(bytes.NewReader(pdf)),
		FileName: fmt.Sprintf("incident-%s.pdf", inc.ID.String()),
		MIME:     "application/pdf",
		Caption:  "📄 Incident report",
	}
	_, err := h.api.Send(chat, document)
	return err
}

func (h *Handler) setSeverity(c telebot.Context, sev incident.Severity) (*incident.Incident, error) {
	ctx, cancel := reqContext()
	defer cancel()

	return h.svc.SetSeverity(ctx, c.Chat().ID, threadID(c), sev)
}
