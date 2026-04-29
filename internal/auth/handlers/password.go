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

func (h *Handler) ForgotPasswordHandler(w http.ResponseWriter, r *http.Request) error {
	var req auth.ChangePasswordStartInput

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn("Ошибка: не удалось прочитать тело запроса: %v", err)
		return apierror.ErrInvalidJSON
	}

	if err := h.AuthService.ForgotPassword(r.Context(), req.Email); err != nil {
		logger.Warn("Ошибка при смене пароля для email: %v: %v", req.Email, err)
		return apierror.Wrap(err, apierror.ErrInternal)
	}

	return httpx.WriteJSON(w, http.StatusOK, map[string]any{"message": fmt.Sprintf("Код подтверждения отправлен на адрес электронной почты: %s", req.Email)})
}

func (h *Handler) VerifyResetCodeHandler(w http.ResponseWriter, r *http.Request) error {
	var req auth.PasswordVerifyInput

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn("Ошибка: не удалось прочитать тело запроса: %v", err)
		return apierror.ErrInvalidJSON
	}

	if err := h.AuthService.VerifyResetCode(r.Context(), req); err != nil {
		logger.Warn("Ошибка при смене пароля для email: %v: %v", req.Email, err)
		return apierror.Wrap(err, apierror.ErrInternal)
	}

	return httpx.WriteJSON(w, http.StatusOK, nil)
}

func (h *Handler) ResetPasswordHandler(w http.ResponseWriter, r *http.Request) error {
	var req auth.ChangePasswordConfirmInput

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn("Ошибка: не удалось прочитать тело запроса: %v", err)
		return apierror.ErrInvalidJSON
	}

	tokens, err := h.AuthService.ResetPassword(r.Context(), req)
	if err != nil {
		logger.Warn("Ошибка при смене пароля для email: %v: %v", req.Email, err)
		return apierror.Wrap(err, apierror.ErrInternal)
	}

	auth.SetRefreshToken(w, tokens.RefreshToken)

	return httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"message":      "Пароль успешно изменён",
		"access_token": tokens.AccessToken,
	})
}
