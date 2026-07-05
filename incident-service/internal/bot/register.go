package bot

import "gopkg.in/telebot.v3"

// Register binds all command and inline-panel handlers to b.
func (h *Handler) Register(b *telebot.Bot) {
	b.Handle("/start", h.HandleStart)
	b.Handle("/incident", h.HandleIncident)
	b.Handle("/timeline", h.HandleTimeline)
	b.Handle("/dashboard", h.HandleDashboard)

	b.Handle(&btnTimeline, h.handleShowTimeline)
	b.Handle(&btnDashboard, h.handleShowDashboard)
	b.Handle(&btnClose, h.handleCloseIncident)
	b.Handle(&btnReopen, h.handleReopenIncident)
	b.Handle(&btnSeverity, h.handleChangeSeverity)
	b.Handle(&btnSevBack, h.handleSeverityBack)

	b.Handle(&btnSevLow, h.handleSetSeverity)

	b.Handle(telebot.OnText, h.HandleTopicText)

	// Any media of any type may be attached to the timeline when media uploads
	// are enabled; albums are recorded together as a single event.
	for _, ev := range []string{
		telebot.OnPhoto,
		telebot.OnVideo,
		telebot.OnVideoNote,
		telebot.OnDocument,
		telebot.OnVoice,
		telebot.OnAudio,
		telebot.OnAnimation,
		telebot.OnSticker,
	} {
		b.Handle(ev, h.HandleTopicMedia)
	}
}
