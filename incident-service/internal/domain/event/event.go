package event

import (
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ID         uuid.UUID
	IncidentID uuid.UUID
	Type       EventType
	UserID     *int64
	Username   string
	Message    string
	MediaURLs  []string
	CreatedAt  time.Time
}

type EventType string

const (
	TypeIncidentCreated  EventType = "INCIDENT_CREATED"
	TypeCommentAdded     EventType = "COMMENT_ADDED"
	TypeIncidentClosed   EventType = "INCIDENT_CLOSED"
	TypeIncidentReopened EventType = "INCIDENT_REOPENED"
	TypeSeverityChanged  EventType = "SEVERITY_CHANGED"
)

// Label is the human-readable name of a lifecycle event. It is empty for
// COMMENT_ADDED, which carries no label of its own.
func (t EventType) Label() string {
	switch t {
	case TypeIncidentCreated:
		return "Incident opened"
	case TypeIncidentReopened:
		return "Incident reopened"
	case TypeIncidentClosed:
		return "Incident closed"
	case TypeSeverityChanged:
		return "Severity changed"
	default:
		return ""
	}
}

// Summary is how the event reads on a timeline: the lifecycle label combined
// with any message the user attached, so status changes never render as blank
// lines and an opened incident is not mistaken for the first comment.
func (e Event) Summary() string {
	label := e.Type.Label()
	switch {
	case label != "" && e.Message != "":
		return label + ": " + e.Message
	case label != "":
		return label
	default:
		return e.Message
	}
}
