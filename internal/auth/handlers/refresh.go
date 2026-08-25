package handlers

import (
	"net/http"

	"github.com/Halturshik/TicketAgregator-API/internal/auth"
	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/common/httpx"
)

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) error {
	refreshToken, err := auth.GetRefreshToken(r)
	if err != nil {
		return apierror.ErrUnauthorized
	}

	tokens, err := h.service.Refresh(r.Context(), refreshToken)
	if err != nil {
		return apierror.Wrap(err, apierror.ErrInternal)
	}

	auth.SetRefreshToken(w, tokens.RefreshToken)

	return httpx.WriteJSON(w, http.StatusOK, map[string]any{"access_token": tokens.AccessToken})
}
