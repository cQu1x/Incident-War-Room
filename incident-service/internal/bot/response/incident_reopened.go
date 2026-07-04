package response

import (
	"fmt"
	"strings"

	"github.com/cQu1x/Incident-War-Room/internal/domain/incident"
)

func IncidentReopened(inc incident.Incident, topicURL string) string {
	var b strings.Builder

	b.WriteString("🔄 <b>Incident reopened</b>\n\n")
	fmt.Fprintf(&b, "<b>Title:</b> %s\n", escape(inc.Title))
	fmt.Fprintf(&b, "<b>Severity:</b> %s\n", severityIcon(inc.Severity))
	fmt.Fprintf(&b, "<b>Status:</b> %s\n", escape(string(inc.Status)))
	fmt.Fprintf(&b, "<b>ID:</b> <code>%s</code>", shortID(inc.ID))

	if topicURL != "" {
		fmt.Fprintf(&b, "\n\n📌 <a href=\"%s\">Open incident topic</a>", escape(topicURL))
	}

	return b.String()
}
