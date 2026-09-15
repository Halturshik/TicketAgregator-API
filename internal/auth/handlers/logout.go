package handlers

import (
	"log/slog"
	"net/http"

	"github.com/Halturshik/TicketAgregator-API/internal/auth"
	"github.com/Halturshik/TicketAgregator-API/internal/common/httpx"
)

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) error {
	refreshToken, _ := auth.GetRefreshToken(r)

	if refreshToken != "" {
		if err := h.service.Logout(r.Context(), refreshToken); err != nil {
			slog.WarnContext(r.Context(), "Не удалось полностью инвалидировать refresh-токен при logout",
				slog.Any("error", err),
			)
		}
	}

	auth.ClearRefreshToken(w)

	return httpx.WriteJSON(w, http.StatusOK, nil)
}
