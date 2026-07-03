// Package auth issues and verifies the short-lived JSON Web Tokens that grant
// access to the web dashboard. A token is minted when an incident is closed and
// embedded in the dashboard link sent to the operator; the HTTP API and the
// frontend both verify it before serving incident data.
package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNoSecret     = errors.New("auth: signing secret is not configured")
	ErrInvalidToken = errors.New("auth: token is malformed or its signature is invalid")
	ErrExpiredToken = errors.New("auth: token has expired")
)

// Claims are the verified contents of a dashboard token.
type Claims struct {
	IncidentID uuid.UUID
	IssuedAt   time.Time
	ExpiresAt  time.Time
}

// Issuer signs and verifies dashboard tokens with a single HMAC-SHA256 secret.
type Issuer struct {
	secret []byte
	ttl    time.Duration
	now    func() time.Time
}

// NewIssuer builds an issuer that signs tokens valid for ttl.
func NewIssuer(secret string, ttl time.Duration) *Issuer {
	return &Issuer{secret: []byte(secret), ttl: ttl, now: time.Now}
}

type header struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

type payload struct {
	Sub string `json:"sub"`
	Iat int64  `json:"iat"`
	Exp int64  `json:"exp"`
}

// Issue mints a token that authorises dashboard access for incidentID.
func (i *Issuer) Issue(incidentID uuid.UUID) (string, error) {
	if len(i.secret) == 0 {
		return "", ErrNoSecret
	}

	issued := i.now().UTC()
	head := encodeSegment(header{Alg: "HS256", Typ: "JWT"})
	body := encodeSegment(payload{
		Sub: incidentID.String(),
		Iat: issued.Unix(),
		Exp: issued.Add(i.ttl).Unix(),
	})

	signing := head + "." + body
	return signing + "." + i.sign(signing), nil
}

// Verify checks a token's signature and expiry and returns its claims.
func (i *Issuer) Verify(token string) (Claims, error) {
	if len(i.secret) == 0 {
		return Claims{}, ErrNoSecret
	}

	head, body, sig, ok := split(token)
	if !ok {
		return Claims{}, ErrInvalidToken
	}

	expected := i.sign(head + "." + body)
	if !hmac.Equal([]byte(sig), []byte(expected)) {
		return Claims{}, ErrInvalidToken
	}

	var p payload
	if err := decodeSegment(body, &p); err != nil {
		return Claims{}, ErrInvalidToken
	}

	id, err := uuid.Parse(p.Sub)
	if err != nil {
		return Claims{}, ErrInvalidToken
	}

	claims := Claims{
		IncidentID: id,
		IssuedAt:   time.Unix(p.Iat, 0).UTC(),
		ExpiresAt:  time.Unix(p.Exp, 0).UTC(),
	}
	if !i.now().UTC().Before(claims.ExpiresAt) {
		return Claims{}, ErrExpiredToken
	}
	return claims, nil
}

func (i *Issuer) sign(signing string) string {
	mac := hmac.New(sha256.New, i.secret)
	mac.Write([]byte(signing))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func encodeSegment(v any) string {
	raw, _ := json.Marshal(v)
	return base64.RawURLEncoding.EncodeToString(raw)
}

func decodeSegment(segment string, v any) error {
	raw, err := base64.RawURLEncoding.DecodeString(segment)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, v)
}

func split(token string) (head, body, sig string, ok bool) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", "", false
	}
	return parts[0], parts[1], parts[2], true
}
