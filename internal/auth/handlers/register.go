package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Halturshik/TicketAgregator-API/internal/auth"
	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/common/httpx"
)

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) error {
	var req auth.RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierror.ErrInvalidJSON
	}

	if err := h.service.StartRegistration(r.Context(), req); err != nil {
		return apierror.Wrap(err, apierror.ErrInternal)
	}
	return httpx.WriteJSON(w, http.StatusOK, map[string]any{"message": fmt.Sprintf("Код подтверждения отправлен на адрес электронной почты: %s", req.Email)})
}

func (h *Handler) ConfirmRegistration(w http.ResponseWriter, r *http.Request) error {
	var req auth.ConfirmRegisterInput

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierror.ErrInvalidJSON
	}

	tokens, err := h.service.ConfirmRegistration(r.Context(), req)
	if err != nil {
		return apierror.Wrap(err, apierror.ErrInternal)
	}

	auth.SetRefreshToken(w, tokens.RefreshToken)
	return httpx.WriteJSON(w, http.StatusCreated, map[string]any{
		"message":      "Вы успешно прошли регистрацию. Хороших поездок!",
		"access_token": tokens.AccessToken,
	})
}
