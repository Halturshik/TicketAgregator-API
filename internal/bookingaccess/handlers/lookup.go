package handlers

import (
	"net/http"

	"github.com/Halturshik/TicketAgregator-API/internal/bookingaccess"
	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/common/httpx"
)

func (h *Handler) Lookup(w http.ResponseWriter, r *http.Request) error {
	input, err := decode[bookingaccess.LookupInput](r)
	if err != nil {
		return err
	}
	output, err := h.service.Lookup(r.Context(), input)
	if err != nil {
		return apierror.Wrap(err, apierror.ErrInternal)
	}
	w.Header().Set("Cache-Control", "no-store")
	return httpx.WriteJSON(w, http.StatusOK, output)
}
