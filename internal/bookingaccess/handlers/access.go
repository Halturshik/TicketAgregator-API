package handlers

import (
	"net/http"

	"github.com/Halturshik/TicketAgregator-API/internal/bookingaccess"
	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/common/httpx"
)

func (h *Handler) RequestAccess(w http.ResponseWriter, r *http.Request) error {
	input, err := decode[bookingaccess.AccessRequestInput](r)
	if err != nil {
		return err
	}
	output, err := h.service.RequestAccess(r.Context(), input)
	if err != nil {
		return apierror.Wrap(err, apierror.ErrInternal)
	}
	w.Header().Set("Cache-Control", "no-store")
	return httpx.WriteJSON(w, http.StatusAccepted, output)
}

func (h *Handler) ConfirmAccess(w http.ResponseWriter, r *http.Request) error {
	input, err := decode[bookingaccess.AccessConfirmInput](r)
	if err != nil {
		return err
	}
	output, err := h.service.ConfirmAccess(r.Context(), input)
	if err != nil {
		return apierror.Wrap(err, apierror.ErrInternal)
	}
	bookingaccess.SetAccessToken(w, output.Token, output.ExpiresAt)
	w.Header().Set("Cache-Control", "no-store")
	return httpx.WriteJSON(w, http.StatusOK, output)
}
