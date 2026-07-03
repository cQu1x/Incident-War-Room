package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/cQu1x/Incident-War-Room/internal/domain/incident"
	"github.com/cQu1x/Incident-War-Room/internal/errs"
)

func TestNewIncidentResponseNilTelegraphURLsSerialiseAsEmptyArray(t *testing.T) {
	inc := incident.Incident{
		ID:        uuid.New(),
		Title:     "db down",
		Severity:  incident.SeverityHigh,
		Status:    incident.StatusActive,
		CreatedAt: time.Now(),
	}

	body, err := json.Marshal(newIncidentResponse(inc))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(body), `"telegraphUrls":[]`) {
		t.Fatalf("expected telegraphUrls to be [], got %s", body)
	}
}

func TestNewIncidentResponseMapsFields(t *testing.T) {
	url := "https://reports.example/r/abc.pdf"
	inc := incident.Incident{
		ID:            uuid.New(),
		Title:         "db down",
		Severity:      incident.SeverityMedium,
		Status:        incident.StatusClosed,
		ChatID:        42,
		TopicID:       7,
		TelegraphURLs: []string{"https://telegra.ph/x"},
		ReportURL:     &url,
	}

	resp := newIncidentResponse(inc)

	if resp.ID != inc.ID.String() || resp.Severity != "MEDIUM" || resp.Status != "CLOSED" {
		t.Fatalf("unexpected header fields: %+v", resp)
	}
	if resp.ChatID != 42 || resp.TopicID != 7 {
		t.Fatalf("unexpected chat/topic: %+v", resp)
	}
	if resp.ReportURL == nil || *resp.ReportURL != url {
		t.Fatalf("report url = %v, want %q", resp.ReportURL, url)
	}
}

func TestNewIncidentResponsesPreservesOrder(t *testing.T) {
	incidents := []incident.Incident{
		{ID: uuid.New(), Title: "first"},
		{ID: uuid.New(), Title: "second"},
	}

	resp := newIncidentResponses(incidents)

	if len(resp) != 2 || resp[0].Title != "first" || resp[1].Title != "second" {
		t.Fatalf("unexpected responses: %+v", resp)
	}
}

func TestStatusForKind(t *testing.T) {
	tests := []struct {
		kind errs.Kind
		want int
	}{
		{errs.KindNotFound, http.StatusNotFound},
		{errs.KindValidation, http.StatusBadRequest},
		{errs.KindConflict, http.StatusConflict},
		{errs.KindUnavailable, http.StatusServiceUnavailable},
		{errs.KindInternal, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		if got := statusForKind(tt.kind); got != tt.want {
			t.Errorf("statusForKind(%v) = %d, want %d", tt.kind, got, tt.want)
		}
	}
}

func TestGetIncidentRejectsMalformedID(t *testing.T) {
	_, issuer, h := newTestServer(t)
	token, err := issuer.Issue(uuid.New())
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}

	rec := get(t, h, "/api/v1/incidents/not-a-uuid", token)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}
