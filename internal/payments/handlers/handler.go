package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Halturshik/TicketAgregator-API/internal/auth"
	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/common/httpx"
	"github.com/Halturshik/TicketAgregator-API/internal/payments"
)

type Handler struct {
	service payments.Service
}

func New(service payments.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Pay(w http.ResponseWriter, r *http.Request) error {
	var req payments.PayInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierror.ErrInvalidJSON
	}
	var userID *int
	if id, ok := auth.UserIDFromContext(r.Context()); ok {
		userID = &id
	}
	out, err := h.service.Pay(r.Context(), userID, req.OrderID, req.GuestPaymentToken)
	if err != nil {
		return apierror.Wrap(err, apierror.ErrInternal)
	}
	return httpx.WriteJSON(w, http.StatusOK, out)
}
