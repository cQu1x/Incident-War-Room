package api

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
)

func TestListIncidentsWithoutChatIdPassesNilFilter(t *testing.T) {
	svc, issuer, h := newTestServer(t)
	token, err := issuer.Issue(uuid.New())
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}

	rec := get(t, h, "/api/v1/incidents", token)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if svc.lastChatID != nil {
		t.Fatalf("chatID filter = %v, want nil", *svc.lastChatID)
	}
}

func TestListIncidentsPassesChatIdFilter(t *testing.T) {
	svc, issuer, h := newTestServer(t)
	token, err := issuer.Issue(uuid.New())
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}

	rec := get(t, h, "/api/v1/incidents?chatId=42", token)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if svc.lastChatID == nil || *svc.lastChatID != 42 {
		t.Fatalf("chatID filter = %v, want 42", svc.lastChatID)
	}
}

func TestListIncidentsRejectsInvalidChatId(t *testing.T) {
	svc, issuer, h := newTestServer(t)
	token, err := issuer.Issue(uuid.New())
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}

	rec := get(t, h, "/api/v1/incidents?chatId=not-a-number", token)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	if svc.calls != 0 {
		t.Fatalf("service was called %d times despite invalid chatId", svc.calls)
	}
}
