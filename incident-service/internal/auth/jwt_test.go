package auth

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

const testSecret = "test-signing-secret"

func fixedIssuer(secret string, ttl time.Duration, now time.Time) *Issuer {
	i := NewIssuer(secret, ttl)
	i.now = func() time.Time { return now }
	return i
}

func TestIssueAndVerifyRoundTrip(t *testing.T) {
	now := time.Date(2026, 7, 3, 12, 0, 0, 0, time.UTC)
	i := fixedIssuer(testSecret, time.Hour, now)
	id := uuid.New()

	token, err := i.Issue(id)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}

	claims, err := i.Verify(token)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if claims.IncidentID != id {
		t.Errorf("IncidentID = %s, want %s", claims.IncidentID, id)
	}
	if !claims.ExpiresAt.Equal(now.Add(time.Hour)) {
		t.Errorf("ExpiresAt = %s, want %s", claims.ExpiresAt, now.Add(time.Hour))
	}
}

func TestVerifyRejectsExpiredToken(t *testing.T) {
	issued := time.Date(2026, 7, 3, 12, 0, 0, 0, time.UTC)
	i := fixedIssuer(testSecret, time.Minute, issued)

	token, err := i.Issue(uuid.New())
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}

	i.now = func() time.Time { return issued.Add(2 * time.Minute) }
	if _, err := i.Verify(token); !errors.Is(err, ErrExpiredToken) {
		t.Fatalf("Verify error = %v, want ErrExpiredToken", err)
	}
}

func TestVerifyRejectsTamperedSignature(t *testing.T) {
	i := NewIssuer(testSecret, time.Hour)
	token, err := i.Issue(uuid.New())
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}

	if _, err := i.Verify(token + "x"); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("Verify error = %v, want ErrInvalidToken", err)
	}
}

func TestVerifyRejectsTokenSignedWithAnotherSecret(t *testing.T) {
	token, err := NewIssuer("secret-a", time.Hour).Issue(uuid.New())
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}

	if _, err := NewIssuer("secret-b", time.Hour).Verify(token); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("Verify error = %v, want ErrInvalidToken", err)
	}
}

func TestVerifyRejectsMalformedToken(t *testing.T) {
	i := NewIssuer(testSecret, time.Hour)
	for _, token := range []string{"", "onlyonepart", "two.parts", "a.b.c.d", "..", "a..c"} {
		if _, err := i.Verify(token); !errors.Is(err, ErrInvalidToken) {
			t.Errorf("Verify(%q) error = %v, want ErrInvalidToken", token, err)
		}
	}
}

func TestIssueAndVerifyRequireSecret(t *testing.T) {
	i := NewIssuer("", time.Hour)
	if _, err := i.Issue(uuid.New()); !errors.Is(err, ErrNoSecret) {
		t.Errorf("Issue error = %v, want ErrNoSecret", err)
	}
	if _, err := i.Verify("a.b.c"); !errors.Is(err, ErrNoSecret) {
		t.Errorf("Verify error = %v, want ErrNoSecret", err)
	}
}
