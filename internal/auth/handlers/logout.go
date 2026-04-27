package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Halturshik/TicketAgregator-API/internal/auth"
	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/common/httpx"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
)

func (api *API) LogoutHandler(w http.ResponseWriter, r *http.Request) error {
	var req auth.LogoutInput

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn("Ошибка: не удалось прочитать тело запроса: %v", err)
		return apierror.ErrInvalidJSON
	}

	if err := api.AuthService.Logout(r.Context(), req.RefreshToken); err != nil {
		logger.Warn("Ошибка при logout: (поступивший токен: %v): %v", req.RefreshToken[:8], err)
		return apierror.Wrap(err, apierror.ErrInternal)
	}

	return httpx.WriteJSON(w, http.StatusOK, map[string]any{"message": "Вы вышли из личного кабинета"})
}
