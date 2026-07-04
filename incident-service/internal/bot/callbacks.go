package bot

import (
	"log"

	"gopkg.in/telebot.v3"

	"github.com/cQu1x/Incident-War-Room/internal/domain/incident"
)

func (h *Handler) handleShowTimeline(c telebot.Context) error {
	if err := c.Respond(); err != nil {
		return err
	}

	ctx, cancel := reqContext()
	defer cancel()

	topicID := threadID(c)

	msg, err := h.renderTimeline(ctx, c.Chat().ID, topicID)
	if err != nil {
		log.Printf("bot: show timeline: %v", err)
		return c.Send(userError(err))
	}

	return c.Send(msg, &telebot.SendOptions{
		ThreadID:  int(topicID),
		ParseMode: telebot.ModeHTML,
	})
}

func (h *Handler) handleCloseIncident(c telebot.Context) error {
	if err := c.Respond(&telebot.CallbackResponse{Text: "Closing incident…"}); err != nil {
		return err
	}

	_, err := h.closeIncident(c)
	return err
}

func (h *Handler) handleReopenIncident(c telebot.Context) error {
	if err := c.Respond(&telebot.CallbackResponse{Text: "Reopening incident…"}); err != nil {
		return err
	}

	return h.reopenIncident(c)
}

func (h *Handler) handleChangeSeverity(c telebot.Context) error {
	if err := c.Respond(); err != nil {
		return err
	}

	return c.Edit(severityMenu())
}

func (h *Handler) handleSetSeverity(c telebot.Context) error {
	sev := incident.Severity(c.Data())
	if err := c.Respond(&telebot.CallbackResponse{Text: "Severity set to " + string(sev)}); err != nil {
		return err
	}

	inc, err := h.setSeverity(c, sev)
	if err != nil {
		log.Printf("bot: set severity: %v", err)
		return c.Send(userError(err))
	}

	if err := c.Edit(incidentCard(inc.Title, inc.Severity, inc.Status, h.mediaEnabled), incidentMenu()); err != nil {
		return err
	}

	h.refreshAnnouncement(c, *inc)
	return nil
}

func (h *Handler) handleSeverityBack(c telebot.Context) error {
	if err := c.Respond(); err != nil {
		return err
	}

	return c.Edit(incidentMenu())
}
