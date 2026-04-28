package handlers

import (
	"net/http"

	"github.com/Halturshik/TicketAgregator-API/internal/auth"
	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/common/httpx"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
)

func (h *API) Refresh(w http.ResponseWriter, r *http.Request) error {
	refreshToken, err := auth.GetRefreshToken(r)
	if err != nil {
		logger.Warn("Ошибка: не удалось прочитать refresh-токен: %v", err)
		return apierror.ErrUnauthorized
	}

	tokens, err := h.AuthService.Refresh(r.Context(), refreshToken)
	if err != nil {
		logger.Warn("Ошибка при обновлении refresh-токена: (поступивший токен: %v): %v", refreshToken[:8], err)
		return apierror.Wrap(err, apierror.ErrInternal)
	}

	auth.SetRefreshToken(w, tokens.RefreshToken)

	return httpx.WriteJSON(w, http.StatusOK, map[string]any{"access_token": tokens.AccessToken})
}
