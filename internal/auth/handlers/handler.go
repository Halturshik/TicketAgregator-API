package handlers

import "github.com/Halturshik/TicketAgregator-API/internal/auth"

type Handler struct {
	service auth.Service
}

func New(service auth.Service) *Handler {
	return &Handler{service: service}
}
