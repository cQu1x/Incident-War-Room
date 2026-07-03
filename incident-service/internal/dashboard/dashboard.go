// Package dashboard builds the personal, tokenised web-dashboard link an
// operator receives for a Telegram chat.
package dashboard

import (
	"net/url"
)

// TokenIssuer mints an access token that authorises dashboard access for a
// Telegram chat. It is satisfied by *auth.Issuer.
type TokenIssuer interface {
	Issue(chatID int64) (string, error)
}

// Linker builds "<baseURL>/dashboard?token=<jwt>" links.
type Linker struct {
	baseURL string
	issuer  TokenIssuer
}

func NewLinker(baseURL string, issuer TokenIssuer) *Linker {
	return &Linker{baseURL: baseURL, issuer: issuer}
}

// Link issues a fresh token for chatID and returns the dashboard URL that
// carries it.
func (l *Linker) Link(chatID int64) (string, error) {
	token, err := l.issuer.Issue(chatID)
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
