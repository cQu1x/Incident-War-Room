package response

import (
	"strings"
	"testing"
	"time"

	"github.com/cQu1x/Incident-War-Room/internal/domain/incident"
	"github.com/cQu1x/Incident-War-Room/internal/domain/report"
	"github.com/google/uuid"
)

func TestIncidentClosedSubMinuteIsHTMLSafe(t *testing.T) {
	created := time.Date(2026, 6, 13, 8, 0, 0, 0, time.UTC)
	closed := created.Add(30 * time.Second)
	inc := incident.Incident{
		ID:        uuid.New(),
		Title:     "DB is down",
		Status:    incident.StatusClosed,
		CreatedAt: created,
		ClosedAt:  &closed,
	}

	got := IncidentClosed(inc, nil, report.Document{}, "")

	if strings.Contains(got, "<1m") {
		t.Errorf("IncidentClosed() leaks a raw '<' that breaks Telegram HTML parse mode: %q", got)
	}
	if !strings.Contains(got, "&lt;1m") {
		t.Errorf("IncidentClosed() = %q, expected escaped duration &lt;1m", got)
	}
}

func TestIncidentClosed(t *testing.T) {
	created := time.Date(2026, 6, 13, 8, 0, 0, 0, time.UTC)
	closed := time.Date(2026, 6, 13, 10, 15, 0, 0, time.UTC)
	inc := incident.Incident{
		ID:        uuid.MustParse("a1b2c3d4-0000-0000-0000-000000000000"),
		Title:     "DB is down",
		Status:    incident.StatusClosed,
		CreatedAt: created,
		ClosedAt:  &closed,
	}

	got := IncidentClosed(inc, nil, report.Document{}, "")

	for _, want := range []string{
		"<b>Incident closed</b>",
		"DB is down",
		"a1b2c3d4",
		"2026-06-13 10:15 UTC",
		"2h 15m",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("IncidentClosed() = %q, missing %q", got, want)
		}
	}
}

func TestIncidentClosedWithoutClosedAt(t *testing.T) {
	inc := incident.Incident{
		ID:    uuid.New(),
		Title: "DB is down",
	}

	got := IncidentClosed(inc, nil, report.Document{}, "")

	if !strings.Contains(got, "just now") {
		t.Errorf("IncidentClosed() = %q, expected 'just now' fallback", got)
	}
	if strings.Contains(got, "Duration:") {
		t.Errorf("IncidentClosed() should omit duration without ClosedAt: %q", got)
	}
}

func TestIncidentClosedTimelineBlock(t *testing.T) {
	inc := incident.Incident{ID: uuid.New(), Title: "DB is down"}

	placeholder := IncidentClosed(inc, nil, report.Document{}, "")
	if !strings.Contains(placeholder, "📋") || !strings.Contains(placeholder, "Timeline") {
		t.Errorf("IncidentClosed() = %q, expected a Timeline block", placeholder)
	}

	withURLs := IncidentClosed(inc, []string{"https://telegra.ph/page-1", "https://telegra.ph/page-2"}, report.Document{}, "")
	for _, want := range []string{"https://telegra.ph/page-1", "https://telegra.ph/page-2"} {
		if !strings.Contains(withURLs, want) {
			t.Errorf("IncidentClosed() = %q, missing timeline url %q", withURLs, want)
		}
	}
}

func TestIncidentClosedReportBlock(t *testing.T) {
	inc := incident.Incident{ID: uuid.New(), Title: "DB is down"}

	withReport := IncidentClosed(inc, nil, report.Document{URL: "https://reports.example/r/abc.pdf"}, "")
	if !strings.Contains(withReport, "📄") || !strings.Contains(withReport, "https://reports.example/r/abc.pdf") {
		t.Errorf("IncidentClosed() = %q, expected a report link", withReport)
	}

	unavailable := IncidentClosed(inc, nil, report.Document{}, "")
	if !strings.Contains(unavailable, "report could not be generated") {
		t.Errorf("IncidentClosed() = %q, expected report-unavailable notice", unavailable)
	}
}

func TestIncidentClosedDashboardBlock(t *testing.T) {
	inc := incident.Incident{ID: uuid.New(), Title: "DB is down"}
	link := "https://incident-war-room.ru/dashboard?token=a.b.c"

	withLink := IncidentClosed(inc, nil, report.Document{}, link)
	if !strings.Contains(withLink, "🔐") || !strings.Contains(withLink, link) {
		t.Errorf("IncidentClosed() = %q, expected a dashboard link", withLink)
	}

	withoutLink := IncidentClosed(inc, nil, report.Document{}, "")
	if strings.Contains(withoutLink, "Dashboard") {
		t.Errorf("IncidentClosed() = %q, should omit dashboard block when link is empty", withoutLink)
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{30 * time.Second, "<1m"},
		{45 * time.Minute, "45m"},
		{2 * time.Hour, "2h"},
		{2*time.Hour + 15*time.Minute, "2h 15m"},
	}

	for _, tt := range tests {
		if got := formatDuration(tt.d); got != tt.want {
			t.Errorf("formatDuration(%v) = %q, want %q", tt.d, got, tt.want)
		}
	}
}
