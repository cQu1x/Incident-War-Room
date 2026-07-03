package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/cQu1x/Incident-War-Room/internal/auth"
	"github.com/cQu1x/Incident-War-Room/internal/domain/event"
	"github.com/cQu1x/Incident-War-Room/internal/domain/incident"
)

type stubService struct{ calls int }

func (s *stubService) ListIncidents(context.Context) ([]incident.Incident, error) {
	s.calls++
	return nil, nil
}
func (s *stubService) GetIncident(context.Context, uuid.UUID) (*incident.Incident, error) {
	s.calls++
	return &incident.Incident{}, nil
}
func (s *stubService) IncidentTimeline(context.Context, uuid.UUID) (*incident.Incident, []event.Event, error) {
	s.calls++
	return &incident.Incident{}, nil, nil
}
func (s *stubService) IncidentImages(context.Context, uuid.UUID) ([]event.Event, error) {
	s.calls++
	return nil, nil
}

func newTestServer(t *testing.T) (*stubService, *auth.Issuer, http.Handler) {
	t.Helper()
	svc := &stubService{}
	issuer := auth.NewIssuer("api-test-secret", time.Hour)
	return svc, issuer, NewServer(svc, issuer, "*").Handler()
}

func get(t *testing.T, h http.Handler, path, token string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestIncidentEndpointRejectsMissingToken(t *testing.T) {
	svc, _, h := newTestServer(t)

	rec := get(t, h, "/api/v1/incidents", "")

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
	if svc.calls != 0 {
		t.Fatalf("service was called %d times despite unauthorized request", svc.calls)
	}
}

func TestIncidentEndpointRejectsInvalidToken(t *testing.T) {
	svc, _, h := newTestServer(t)

	rec := get(t, h, "/api/v1/incidents", "not-a-real-token")

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
	if svc.calls != 0 {
		t.Fatalf("service was called for an invalid token")
	}
}

func TestIncidentEndpointAcceptsValidToken(t *testing.T) {
	svc, issuer, h := newTestServer(t)
	token, err := issuer.Issue(uuid.New())
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}

	rec := get(t, h, "/api/v1/incidents", token)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if svc.calls != 1 {
		t.Fatalf("service calls = %d, want 1", svc.calls)
	}
}

func TestVerifyEndpointReturnsClaims(t *testing.T) {
	_, issuer, h := newTestServer(t)
	id := uuid.New()
	token, err := issuer.Issue(id)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}

	rec := get(t, h, "/api/v1/auth/verify", token)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var body verifyResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.IncidentID != id.String() {
		t.Fatalf("incidentId = %s, want %s", body.IncidentID, id)
	}
}

func TestVerifyEndpointRejectsMissingToken(t *testing.T) {
	_, _, h := newTestServer(t)

	rec := get(t, h, "/api/v1/auth/verify", "")

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}
