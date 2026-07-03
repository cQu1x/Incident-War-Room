package bot

import (
	"log"

	"gopkg.in/telebot.v3"

	"github.com/cQu1x/Incident-War-Room/internal/bot/response"
)

// HandleDashboard replies to the /dashboard command with a fresh personal
// dashboard link carrying a token scoped to the current chat.
func (h *Handler) HandleDashboard(c telebot.Context) error {
	return h.sendDashboardLink(c)
}

// handleShowDashboard answers the incident-card "Dashboard" button with the
// same personal link. Each press mints a new token.
func (h *Handler) handleShowDashboard(c telebot.Context) error {
	if err := c.Respond(); err != nil {
		return err
	}
	return h.sendDashboardLink(c)
}

func (h *Handler) sendDashboardLink(c telebot.Context) error {
	if h.dashboard == nil {
		return c.Send(response.DashboardUnavailable())
	}

	link, err := h.dashboard.Link(c.Chat().ID)
	if err != nil {
		log.Printf("bot: build dashboard link: %v", err)
		return c.Send(response.DashboardUnavailable())
	}

	return c.Send(response.DashboardLink(link), &telebot.SendOptions{
		ThreadID:              int(threadID(c)),
		ParseMode:             telebot.ModeHTML,
		DisableWebPagePreview: true,
	})
}
