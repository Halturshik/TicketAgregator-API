package handlers

import (
	"net/http"

	"github.com/Halturshik/TicketAgregator-API/internal/auth"
	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/common/httpx"
)

func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) error {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		return apierror.ErrUnauthorized
	}

	profile, err := h.service.GetProfile(r.Context(), userID)
	if err != nil {
		return apierror.Wrap(err, apierror.ErrInternal)
	}

	return httpx.WriteJSON(w, http.StatusOK, profile)
}
