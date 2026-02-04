package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Halturshik/TicketAgregator-API/internal/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/auth"
	"github.com/Halturshik/TicketAgregator-API/internal/logger"
)

func (api *API) RegisterHandler(w http.ResponseWriter, r *http.Request) error {
	var req auth.RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn("Ошибка: не удалось прочитать тело запроса: %v", err)
		return apierror.ErrInvalidJSON
	}

	if err := api.AuthService.StartRegistration(r.Context(), req); err != nil {
		logger.Warn("Ошибка при регистрации: %v", err)
		return apierror.Wrap(err, apierror.ErrInternal)
	}

	writeJSON(w, http.StatusCreated, map[string]any{"message": fmt.Sprintf("Код подтверждения отправлен на адрес электронной почты: %s", req.Email)})
}

func (api *API) ConfirmRegistrationHandler(w http.ResponseWriter, r *http.Request) error {
	var req struct {
		Code string
		auth.RegisterInput
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn("Ошибка: не удалось прочитать тело запроса: %v", err)
		return apierror.ErrInvalidJSON
	}

	if err := api.AuthService.ConfirmRegistration(r.Context(), req.RegisterInput, req.Code); err != nil {
		logger.Warn("Ошибка при регистрации: %v", err)
		return apierror.Wrap(err, apierror.ErrInternal)
	}

	writeJSON(w, http.StatusCreated, map[string]any{"message": fmt.Sprintf("%s, вы успешно прошли регистрацию. Хороших поездок!", req.FirstName)})
}
