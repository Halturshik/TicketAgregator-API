package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Halturshik/TicketAgregator-API/internal/auth"
	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/common/httpx"
	"github.com/Halturshik/TicketAgregator-API/internal/users"
)

func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) error {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		return apierror.ErrUnauthorized
	}

	var req users.UpdateProfileInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierror.ErrInvalidJSON
	}

	profile, err := h.service.UpdateProfile(r.Context(), userID, req)
	if err != nil {
		return apierror.Wrap(err, apierror.ErrInternal)
	}

	return httpx.WriteJSON(w, http.StatusOK, profile)
}
