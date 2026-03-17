package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Halturshik/TicketAgregator-API/internal/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/auth/types"
	"github.com/Halturshik/TicketAgregator-API/internal/httpx"
	"github.com/Halturshik/TicketAgregator-API/internal/logger"
)

func (api *API) LoginStartHandler(w http.ResponseWriter, r *http.Request) error {
	var req types.LoginStartInput

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn("Ошибка: не удалось прочитать тело запроса: %v", err)
		return apierror.ErrInvalidJSON
	}

	if err := api.AuthService.LoginStart(r.Context(), req); err != nil {
		logger.Warn("Ошибка при старте авторизации для email: %v : %v", req.Email, err)
		return apierror.Wrap(err, apierror.ErrInternal)
	}

	return httpx.WriteJSON(w, http.StatusOK, map[string]any{"message": fmt.Sprintf("Код подтверждения отправлен на адрес электронной почты: %s", req.Email)})
}

func (api *API) LoginConfirmHandler(w http.ResponseWriter, r *http.Request) error {
	var req types.LoginConfirmInput

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn("Ошибка: не удалось прочитать тело запроса: %v", err)
		return apierror.ErrInvalidJSON
	}

	tokens, err := api.AuthService.LoginConfirm(r.Context(), req)
	if err != nil {
		logger.Warn("Ошибка при подтверждении авторизации для email: %v : %v", req.Email, err)
		return apierror.Wrap(err, apierror.ErrInternal)
	}

	logger.Info("Успешная авторизация для email: %s", req.Email)
	return httpx.WriteJSON(w, http.StatusOK, tokens)
}
