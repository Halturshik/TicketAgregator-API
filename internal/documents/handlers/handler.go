package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Halturshik/TicketAgregator-API/internal/auth"
	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/common/httpx"
	"github.com/Halturshik/TicketAgregator-API/internal/documents"
)

type Handler struct {
	service documents.Service
}

func New(service documents.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) error {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		return apierror.ErrUnauthorized
	}

	var passengerID *int
	if raw := r.URL.Query().Get("passenger_id"); raw != "" {
		id, err := strconv.Atoi(raw)
		if err != nil || id <= 0 {
			return apierror.ErrInvalidRequest
		}
		passengerID = &id
	}

	items, err := h.service.List(r.Context(), userID, passengerID)
	if err != nil {
		return apierror.Wrap(err, apierror.ErrInternal)
	}
	return httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) error {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		return apierror.ErrUnauthorized
	}

	var req documents.SaveDocumentInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierror.ErrInvalidJSON
	}
	doc, err := h.service.Create(r.Context(), userID, req)
	if err != nil {
		return apierror.Wrap(err, apierror.ErrInternal)
	}
	return httpx.WriteJSON(w, http.StatusCreated, doc)
}
