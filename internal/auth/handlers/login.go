package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Halturshik/TicketAgregator-API/internal/auth"
	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/common/httpx"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
)

func (h *API) LoginStartHandler(w http.ResponseWriter, r *http.Request) error {
	var req auth.LoginStartInput

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn("Ошибка: не удалось прочитать тело запроса: %v", err)
		return apierror.ErrInvalidJSON
	}

	if err := h.AuthService.LoginStart(r.Context(), req); err != nil {
		logger.Warn("Ошибка при старте авторизации для email: %v : %v", req.Email, err)
		return apierror.Wrap(err, apierror.ErrInternal)
	}

	return httpx.WriteJSON(w, http.StatusOK, map[string]any{"message": fmt.Sprintf("Код подтверждения отправлен на адрес электронной почты: %s", req.Email)})
}

func (h *API) LoginConfirmHandler(w http.ResponseWriter, r *http.Request) error {
	var req auth.LoginConfirmInput

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn("Ошибка: не удалось прочитать тело запроса: %v", err)
		return apierror.ErrInvalidJSON
	}

	tokens, err := h.AuthService.LoginConfirm(r.Context(), req)
	if err != nil {
		logger.Warn("Ошибка при подтверждении авторизации для email: %v : %v", req.Email, err)
		return apierror.Wrap(err, apierror.ErrInternal)
	}

	logger.Info("Успешная авторизация для email: %s", req.Email)

	auth.SetRefreshToken(w, tokens.RefreshToken)
	return httpx.WriteJSON(w, http.StatusOK, map[string]any{"access_token": tokens.AccessToken})
}
