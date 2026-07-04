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
	MediaURL   *string
	CreatedAt  time.Time
}

type EventType string

const (
	TypeIncidentCreated  EventType = "INCIDENT_CREATED"
	TypeCommentAdded     EventType = "COMMENT_ADDED"
	TypeIncidentClosed   EventType = "INCIDENT_CLOSED"
	TypeIncidentReopened EventType = "INCIDENT_REOPENED"
)
