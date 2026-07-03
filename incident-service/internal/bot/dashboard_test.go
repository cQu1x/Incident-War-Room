package bot

import (
	"errors"
	"strings"
	"testing"
)

type stubLinker struct {
	link       string
	err        error
	lastChatID int64
}

func (s *stubLinker) Link(chatID int64) (string, error) {
	s.lastChatID = chatID
	return s.link, s.err
}

func TestHandleDashboardSendsPersonalLink(t *testing.T) {
	linker := &stubLinker{link: "https://incident-war-room.ru/dashboard?token=a.b.c"}
	h := New(&fakeService{}, newFakeAPI(), WithDashboard(linker))
	ctx := &mockContext{chatID: -1001234567890}

	if err := h.HandleDashboard(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if linker.lastChatID != -1001234567890 {
		t.Errorf("linked chatID = %d, want -1001234567890", linker.lastChatID)
	}
	sentContains(t, ctx, "https://incident-war-room.ru/dashboard?token=a.b.c")
}

func TestHandleDashboardWithoutLinker(t *testing.T) {
	h := New(&fakeService{}, newFakeAPI())
	ctx := &mockContext{chatID: 1}

	if err := h.HandleDashboard(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := lastSent(t, ctx); strings.Contains(got, "token") {
		t.Errorf("expected no link when dashboard is unset, got %q", got)
	}
}

func TestHandleDashboardReportsIssuerError(t *testing.T) {
	linker := &stubLinker{err: errors.New("no secret")}
	h := New(&fakeService{}, newFakeAPI(), WithDashboard(linker))
	ctx := &mockContext{chatID: 1}

	if err := h.HandleDashboard(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := lastSent(t, ctx); strings.Contains(got, "token") {
		t.Errorf("expected no link on issuer error, got %q", got)
	}
}
