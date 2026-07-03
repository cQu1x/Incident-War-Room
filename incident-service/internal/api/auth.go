package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/cQu1x/Incident-War-Room/internal/auth"
)

// requireAuth wraps a handler so it only runs when the request carries a valid
// dashboard token in the "Authorization: Bearer <token>" header.
func (s *Server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := s.authenticate(w, r); !ok {
			return
		}
		next(w, r)
	}
}

func (s *Server) authenticate(w http.ResponseWriter, r *http.Request) (auth.Claims, bool) {
	token, ok := bearerToken(r)
	if !ok {
		writeUnauthorized(w, "missing bearer token")
		return auth.Claims{}, false
	}

	claims, err := s.tokens.Verify(token)
	if err != nil {
		writeUnauthorized(w, "invalid or expired token")
		return auth.Claims{}, false
	}
	return claims, true
}

type verifyResponse struct {
	IncidentID string    `json:"incidentId"`
	ExpiresAt  time.Time `json:"expiresAt"`
}

// verifyToken lets the frontend confirm a dashboard token before it renders,
// returning the incident the token was minted for and its expiry.
func (s *Server) verifyToken(w http.ResponseWriter, r *http.Request) {
	claims, ok := s.authenticate(w, r)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, verifyResponse{
		IncidentID: claims.IncidentID.String(),
		ExpiresAt:  claims.ExpiresAt,
	})
}

func bearerToken(r *http.Request) (string, bool) {
	header := r.Header.Get("Authorization")
	const prefix = "Bearer "
	if len(header) <= len(prefix) || !strings.EqualFold(header[:len(prefix)], prefix) {
		return "", false
	}
	token := strings.TrimSpace(header[len(prefix):])
	return token, token != ""
}

func writeUnauthorized(w http.ResponseWriter, message string) {
	writeJSON(w, http.StatusUnauthorized, errorResponse{Error: message})
}
