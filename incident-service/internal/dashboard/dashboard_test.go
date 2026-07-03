package dashboard

import (
	"errors"
	"net/url"
	"testing"

	"github.com/google/uuid"
)

type stubIssuer struct {
	token string
	err   error
}

func (s stubIssuer) Issue(uuid.UUID) (string, error) { return s.token, s.err }

func TestLinkCarriesTokenOnDashboardPath(t *testing.T) {
	linker := NewLinker("https://incident-war-room.ru", stubIssuer{token: "a.b.c"})

	link, err := linker.Link(uuid.New())
	if err != nil {
		t.Fatalf("Link: %v", err)
	}

	u, err := url.Parse(link)
	if err != nil {
		t.Fatalf("parse link %q: %v", link, err)
	}
	if u.Host != "incident-war-room.ru" || u.Path != "/dashboard" {
		t.Errorf("link = %q, want host incident-war-room.ru path /dashboard", link)
	}
	if got := u.Query().Get("token"); got != "a.b.c" {
		t.Errorf("token = %q, want a.b.c", got)
	}
}

func TestLinkEscapesToken(t *testing.T) {
	linker := NewLinker("https://incident-war-room.ru", stubIssuer{token: "a b&c"})

	link, err := linker.Link(uuid.New())
	if err != nil {
		t.Fatalf("Link: %v", err)
	}

	u, _ := url.Parse(link)
	if got := u.Query().Get("token"); got != "a b&c" {
		t.Errorf("round-tripped token = %q, want %q", got, "a b&c")
	}
}

func TestLinkPropagatesIssuerError(t *testing.T) {
	sentinel := errors.New("no secret")
	linker := NewLinker("https://incident-war-room.ru", stubIssuer{err: sentinel})

	if _, err := linker.Link(uuid.New()); !errors.Is(err, sentinel) {
		t.Fatalf("Link error = %v, want %v", err, sentinel)
	}
}
