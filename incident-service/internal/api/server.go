// Package api exposes the incident war room over HTTP so the web frontend can
// read incidents, their timeline and related assets. It is a thin transport
// layer: handlers translate requests into service calls and render domain
// models as JSON, leaving all business logic in the service package.
package api

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/cQu1x/Incident-War-Room/internal/domain/event"
	"github.com/cQu1x/Incident-War-Room/internal/domain/incident"
)

const requestTimeout = 30 * time.Second

// IncidentService is the subset of the service layer the HTTP API depends on.
type IncidentService interface {
	ListIncidents(ctx context.Context, chatID *int64) ([]incident.Incident, error)
	GetIncident(ctx context.Context, id uuid.UUID) (*incident.Incident, error)
	IncidentTimeline(ctx context.Context, id uuid.UUID) (*incident.Incident, []event.Event, error)
	IncidentImages(ctx context.Context, id uuid.UUID) ([]event.Event, error)
}

// Route is an additional handler mounted on the server, identified by a
// net/http ServeMux pattern (e.g. "POST /webhooks/alertmanager").
type Route struct {
	Pattern string
	Handler http.Handler
}

type Server struct {
	svc           IncidentService
	allowedOrigin string
	routes        []Route
}

func NewServer(svc IncidentService, allowedOrigin string, routes ...Route) *Server {
	return &Server{svc: svc, allowedOrigin: allowedOrigin, routes: routes}
}

// Handler builds the HTTP router with all routes and middleware applied.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/incidents", s.listIncidents)
	mux.HandleFunc("GET /api/v1/incidents/{id}", s.getIncident)
	mux.HandleFunc("GET /api/v1/incidents/{id}/timeline", s.incidentTimeline)
	mux.HandleFunc("GET /api/v1/incidents/{id}/images", s.incidentImages)

	for _, route := range s.routes {
		mux.Handle(route.Pattern, route.Handler)
	}

	return s.cors(mux)
}

// Run starts the HTTP server and blocks until it stops.
func (s *Server) Run(addr string) error {
	srv := &http.Server{
		Addr:              addr,
		Handler:           s.Handler(),
		ReadHeaderTimeout: requestTimeout,
	}
	return srv.ListenAndServe()
}

func (s *Server) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", s.allowedOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) context(r *http.Request) (context.Context, context.CancelFunc) {
	return context.WithTimeout(r.Context(), requestTimeout)
}
