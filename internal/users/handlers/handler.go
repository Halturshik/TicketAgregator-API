package handlers

import (
	"github.com/Halturshik/TicketAgregator-API/internal/users"
)

type Handler struct {
	service users.Service
}

func New(service users.Service) *Handler {
	return &Handler{service: service}
}
