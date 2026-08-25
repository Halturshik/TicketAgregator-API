package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Halturshik/TicketAgregator-API/internal/auth"
	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/common/httpx"
)

func (h *Handler) ForgotPassword(w http.ResponseWriter, r *http.Request) error {
	var req auth.ChangePasswordStartInput

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierror.ErrInvalidJSON
	}

	if err := h.service.ForgotPassword(r.Context(), req.Email); err != nil {
		return apierror.Wrap(err, apierror.ErrInternal)
	}

	return httpx.WriteJSON(w, http.StatusOK, map[string]any{"message": fmt.Sprintf("Код подтверждения отправлен на адрес электронной почты: %s", req.Email)})
}

func (h *Handler) VerifyResetCode(w http.ResponseWriter, r *http.Request) error {
	var req auth.PasswordVerifyInput

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierror.ErrInvalidJSON
	}

	if err := h.service.VerifyResetCode(r.Context(), req); err != nil {
		return apierror.Wrap(err, apierror.ErrInternal)
	}

	return httpx.WriteJSON(w, http.StatusOK, nil)
}

func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) error {
	var req auth.ChangePasswordConfirmInput

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierror.ErrInvalidJSON
	}

	tokens, err := h.service.ResetPassword(r.Context(), req)
	if err != nil {
		return apierror.Wrap(err, apierror.ErrInternal)
	}

	auth.SetRefreshToken(w, tokens.RefreshToken)

	return httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"message":      "Пароль успешно изменён",
		"access_token": tokens.AccessToken,
	})
}
