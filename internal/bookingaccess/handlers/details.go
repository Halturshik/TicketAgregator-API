package handlers

import (
	"net/http"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/common/httpx"
	"github.com/go-chi/chi/v5"
)

func (h *Handler) Details(w http.ResponseWriter, r *http.Request) error {
	userID, token := accessCredentials(r)
	output, err := h.service.Details(
		r.Context(), userID, chi.URLParam(r, "orderNumber"), token,
	)
	if err != nil {
		return apierror.Wrap(err, apierror.ErrInternal)
	}
	w.Header().Set("Cache-Control", "no-store")
	return httpx.WriteJSON(w, http.StatusOK, output)
}
