package response

import "fmt"

// DashboardLink renders the message carrying a personal dashboard link.
func DashboardLink(url string) string {
	return fmt.Sprintf("🔐 <b>Dashboard</b>\n<a href=\"%s\">Open your incident dashboard</a>", escape(url))
}

// DashboardUnavailable is shown when no dashboard link can be issued, usually
// because the signing secret is not configured.
func DashboardUnavailable() string {
	return "The dashboard is not available right now."
}
