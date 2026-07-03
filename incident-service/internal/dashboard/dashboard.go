// Package dashboard turns an incident into the tokenised web-dashboard link
// that operators receive when an incident is closed.
package dashboard

import (
	"net/url"

	"github.com/google/uuid"
)

// TokenIssuer mints an access token that authorises dashboard access for an
// incident. It is satisfied by *auth.Issuer.
type TokenIssuer interface {
	Issue(incidentID uuid.UUID) (string, error)
}

// Linker builds "<baseURL>/dashboard?token=<jwt>" links.
type Linker struct {
	baseURL string
	issuer  TokenIssuer
}

func NewLinker(baseURL string, issuer TokenIssuer) *Linker {
	return &Linker{baseURL: baseURL, issuer: issuer}
}

// Link issues a fresh token for incidentID and returns the dashboard URL that
// carries it.
func (l *Linker) Link(incidentID uuid.UUID) (string, error) {
	token, err := l.issuer.Issue(incidentID)
	if err != nil {
		return "", err
	}

	u, err := url.Parse(l.baseURL)
	if err != nil {
		return "", err
	}
	u.Path, err = url.JoinPath(u.Path, "dashboard")
	if err != nil {
		return "", err
	}
	u.RawQuery = url.Values{"token": {token}}.Encode()
	return u.String(), nil
}
