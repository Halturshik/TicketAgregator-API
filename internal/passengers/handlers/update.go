package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Halturshik/TicketAgregator-API/internal/auth"
	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/common/httpx"
	"github.com/Halturshik/TicketAgregator-API/internal/passengers"
	"github.com/go-chi/chi/v5"
)

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) error {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		return apierror.ErrUnauthorized
	}
	passengerID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || passengerID <= 0 {
		return apierror.ErrInvalidRequest
	}
	var req passengers.SavePassengerInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierror.ErrInvalidJSON
	}
	item, err := h.service.Update(r.Context(), userID, passengerID, req)
	if err != nil {
		return apierror.Wrap(err, apierror.ErrInternal)
	}
	return httpx.WriteJSON(w, http.StatusOK, item)
}
