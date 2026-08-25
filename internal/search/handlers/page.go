package handlers

import (
	"net/http"
	"strconv"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/common/httpx"
	"github.com/go-chi/chi/v5"
)

func (h *Handler) GetPage(w http.ResponseWriter, r *http.Request) error {
	searchID := chi.URLParam(r, "id")
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	result, err := h.service.GetPage(r.Context(), searchID, offset, limit, optionalUserID(r))
	if err != nil {
		return apierror.Wrap(err, apierror.ErrInternal)
	}
	return httpx.WriteJSON(w, http.StatusOK, result)
}
