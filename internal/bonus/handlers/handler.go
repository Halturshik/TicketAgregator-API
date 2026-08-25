package handlers

import (
	"github.com/Halturshik/TicketAgregator-API/internal/bonus"
)

type Handler struct {
	service bonus.Service
}

func New(service bonus.Service) *Handler {
	return &Handler{service: service}
}
