package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/common/httpx"
	"github.com/Halturshik/TicketAgregator-API/internal/trips"
)

func (h *Handler) Lookup(w http.ResponseWriter, r *http.Request) error {
	var input trips.LookupInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		return apierror.ErrInvalidJSON
	}
	result, err := h.service.Lookup(r.Context(), input)
	if err != nil {
		return apierror.Wrap(err, apierror.ErrInternal)
	}
	w.Header().Set("Cache-Control", "no-store")
	return httpx.WriteJSON(w, http.StatusOK, result)
}
