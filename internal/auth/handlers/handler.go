package handlers

import "github.com/Halturshik/TicketAgregator-API/internal/auth"

type Handler struct {
	AuthService auth.AuthService
}

func New(authService auth.AuthService) *Handler {
	return &Handler{
		AuthService: authService,
	}
}
