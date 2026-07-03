package bot

import "gopkg.in/telebot.v3"

var (
	btnTimeline  = telebot.Btn{Unique: "show_timeline", Text: "📜 Show Timeline"}
	btnClose     = telebot.Btn{Unique: "close_incident", Text: "✅ Close Incident"}
	btnSeverity  = telebot.Btn{Unique: "change_severity", Text: "⚠️ Change Severity"}
	btnDashboard = telebot.Btn{Unique: "show_dashboard", Text: "🔐 Dashboard"}

	btnSevLow    = telebot.Btn{Unique: "set_severity", Text: "🟢 Low", Data: "LOW"}
	btnSevMedium = telebot.Btn{Unique: "set_severity", Text: "🟡 Medium", Data: "MEDIUM"}
	btnSevHigh   = telebot.Btn{Unique: "set_severity", Text: "🔴 High", Data: "HIGH"}
	btnSevBack   = telebot.Btn{Unique: "severity_back", Text: "⬅️ Back"}
)

func incidentMenu() *telebot.ReplyMarkup {
	m := &telebot.ReplyMarkup{}
	m.Inline(
		m.Row(btnTimeline, btnDashboard),
		m.Row(btnClose, btnSeverity),
	)
	return m
}

func severityMenu() *telebot.ReplyMarkup {
	m := &telebot.ReplyMarkup{}
	m.Inline(
		m.Row(btnSevLow, btnSevMedium, btnSevHigh),
		m.Row(btnSevBack),
	)
	return m
}
