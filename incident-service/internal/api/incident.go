package api

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"github.com/cQu1x/Incident-War-Room/internal/errs"
)

func (s *Server) listIncidents(w http.ResponseWriter, r *http.Request) {
	chatID, err := chatIDFilter(r)
	if err != nil {
		writeError(w, err)
		return
	}

	ctx, cancel := s.context(r)
	defer cancel()

	incidents, err := s.svc.ListIncidents(ctx, chatID)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, newIncidentResponses(incidents))
}

// chatIDFilter reads the optional "chatId" query parameter. It returns nil when
// the parameter is absent, or a validation error when it is not a valid int64.
func chatIDFilter(r *http.Request) (*int64, error) {
	raw := r.URL.Query().Get("chatId")
	if raw == "" {
		return nil, nil
	}

	chatID, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return nil, errs.New(errs.KindValidation, "api.chatIDFilter", "invalid chatId")
	}
	return &chatID, nil
}

func (s *Server) getIncident(w http.ResponseWriter, r *http.Request) {
	id, err := incidentID(r)
	if err != nil {
		writeError(w, err)
		return
	}

	ctx, cancel := s.context(r)
	defer cancel()

	inc, err := s.svc.GetIncident(ctx, id)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, newIncidentResponse(*inc))
}

func (s *Server) incidentTimeline(w http.ResponseWriter, r *http.Request) {
	id, err := incidentID(r)
	if err != nil {
		writeError(w, err)
		return
	}

	ctx, cancel := s.context(r)
	defer cancel()

	inc, events, err := s.svc.IncidentTimeline(ctx, id)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, newTimelineResponse(*inc, events))
}

// incidentImages serves the images attached to an incident's timeline: every
// event that carries an uploaded photo, with its caption and author. Returns
// 404 when the incident does not exist.
func (s *Server) incidentImages(w http.ResponseWriter, r *http.Request) {
	id, err := incidentID(r)
	if err != nil {
		writeError(w, err)
		return
	}

	ctx, cancel := s.context(r)
	defer cancel()

	images, err := s.svc.IncidentImages(ctx, id)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, newImageResponses(images))
}

func incidentID(r *http.Request) (uuid.UUID, error) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		return uuid.Nil, errs.New(errs.KindValidation, "api.incidentID", "invalid incident id")
	}
	return id, nil
}
