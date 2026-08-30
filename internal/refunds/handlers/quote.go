package handlers

import (
	"net/http"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/common/httpx"
)

func (h *Handler) Quote(w http.ResponseWriter, r *http.Request) error {
	orderID, userID, input, err := parseRequest(r)
	if err != nil {
		return err
	}
	output, err := h.service.Quote(r.Context(), userID, orderID, input)
	if err != nil {
		return apierror.Wrap(err, apierror.ErrInternal)
	}
	return httpx.WriteJSON(w, http.StatusOK, output)
}
