package bot

import (
	"log"

	"gopkg.in/telebot.v3"

	"github.com/cQu1x/Incident-War-Room/internal/bot/response"
)

// HandleDashboard replies with a fresh personal dashboard link carrying a token
// scoped to the current chat. Each invocation mints a new token.
func (h *Handler) HandleDashboard(c telebot.Context) error {
	if h.dashboard == nil {
		return c.Send(response.DashboardUnavailable())
	}

	link, err := h.dashboard.Link(c.Chat().ID)
	if err != nil {
		log.Printf("bot: build dashboard link: %v", err)
		return c.Send(response.DashboardUnavailable())
	}

	return c.Send(response.DashboardLink(link), telebot.ModeHTML, telebot.NoPreview)
}
