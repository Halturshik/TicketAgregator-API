package handlers

import (
	"github.com/Halturshik/TicketAgregator-API/internal/passengers"
)

type Handler struct {
	service passengers.Service
}

func New(service passengers.Service) *Handler {
	return &Handler{service: service}
}
