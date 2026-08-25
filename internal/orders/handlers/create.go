package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Halturshik/TicketAgregator-API/internal/auth"
	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/common/httpx"
	"github.com/Halturshik/TicketAgregator-API/internal/orders"
)

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
