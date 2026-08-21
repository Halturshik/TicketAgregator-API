package handlers

import (
	"github.com/Halturshik/TicketAgregator-API/internal/search"
)

type Handler struct {
	service search.Service
}

func New(service search.Service) *Handler {
	return &Handler{service: service}
}
