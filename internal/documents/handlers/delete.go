package handlers

import (
	"net/http"
	"strconv"

	"github.com/Halturshik/TicketAgregator-API/internal/auth"
	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/go-chi/chi/v5"
)

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) error {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		return apierror.ErrUnauthorized
	}
	documentID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || documentID <= 0 {
		return apierror.ErrInvalidRequest
	}
	if err := h.service.Delete(r.Context(), userID, documentID); err != nil {
		return apierror.Wrap(err, apierror.ErrInternal)
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}
