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

func (api *API) RegisterHandler(w http.ResponseWriter, r *http.Request) error {
	var req types.RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn("Ошибка: не удалось прочитать тело запроса: %v", err)
		return apierror.ErrInvalidJSON
	}

	if err := api.AuthService.StartRegistration(r.Context(), req); err != nil {
		logger.Warn("Ошибка при регистрации: %v", err)
		return apierror.Wrap(err, apierror.ErrInternal)
	}
	return httpx.WriteJSON(w, http.StatusOK, map[string]any{"message": fmt.Sprintf("Код подтверждения отправлен на адрес электронной почты: %s", req.Email)})
}

func (api *API) ConfirmRegistrationHandler(w http.ResponseWriter, r *http.Request) error {
	var req types.ConfirmRegisterInput

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn("Ошибка: не удалось прочитать тело запроса: %v", err)
		return apierror.ErrInvalidJSON
	}

	if err := api.AuthService.ConfirmRegistration(r.Context(), req); err != nil {
		logger.Warn("Ошибка при регистрации: %v", err)
		return apierror.Wrap(err, apierror.ErrInternal)
	}

	logger.Info("Успешная регистрация для email: %s", req.Email)
	return httpx.WriteJSON(w, http.StatusCreated, map[string]any{"message": fmt.Sprintf("Вы успешно прошли регистрацию. Хороших поездок!")})
}
