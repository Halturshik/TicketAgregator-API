package handlers

import (
	"net/http"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/common/httpx"
)

func (h *Handler) ListCities(w http.ResponseWriter, r *http.Request) error {
	result, err := h.service.ListCities(r.Context())
	if err != nil {
		return apierror.Wrap(err, apierror.ErrInternal)
	}
	return httpx.WriteJSON(w, http.StatusOK, result)
}
