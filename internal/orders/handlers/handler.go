package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Halturshik/TicketAgregator-API/internal/auth"
	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/common/httpx"
	"github.com/Halturshik/TicketAgregator-API/internal/orders"
)

type Handler struct {
	service orders.Service
}

func New(service orders.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) error {
	var req orders.CreateOrderInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierror.ErrInvalidJSON
	}
	var userID *int
	if id, ok := auth.UserIDFromContext(r.Context()); ok {
		userID = &id
	}
	order, err := h.service.Create(r.Context(), userID, req)
	if err != nil {
		return apierror.Wrap(err, apierror.ErrInternal)
	}
	return httpx.WriteJSON(w, http.StatusCreated, order)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) error {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		return apierror.ErrUnauthorized
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	items, err := h.service.List(r.Context(), userID, orders.ListFilter{
		Transport: r.URL.Query().Get("transport"),
		Limit:     limit,
		Offset:    offset,
	})
	if err != nil {
		return apierror.Wrap(err, apierror.ErrInternal)
	}
	return httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}
