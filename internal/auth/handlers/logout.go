package handlers

import (
	"net/http"

	"github.com/Halturshik/TicketAgregator-API/internal/auth"
	"github.com/Halturshik/TicketAgregator-API/internal/common/httpx"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
)

func (h *Handler) LogoutHandler(w http.ResponseWriter, r *http.Request) error {
	refreshToken, _ := auth.GetRefreshToken(r)

	if refreshToken != "" {
		if err := h.AuthService.Logout(r.Context(), refreshToken); err != nil {
			logger.Warn("Не удалось полностью инвалидировать refresh-токен при logout: %v", err)
		}
	}

	auth.ClearRefreshToken(w)

	return httpx.WriteJSON(w, http.StatusOK, nil)
}
