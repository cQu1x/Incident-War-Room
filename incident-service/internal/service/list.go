package service

import (
	"context"

	"github.com/cQu1x/Incident-War-Room/internal/domain/incident"
)

// ListIncidents returns incidents ordered from newest to oldest. When chatID
// is non-nil, only incidents belonging to that Telegram chat are returned.
func (s *Service) ListIncidents(ctx context.Context, chatID *int64) ([]incident.Incident, error) {
	return s.incidents.List(ctx, chatID)
}
