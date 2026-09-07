package handlers

import "github.com/Halturshik/TicketAgregator-API/internal/trips"

type Handler struct {
	service trips.Service
}

func New(service trips.Service) *Handler {
	return &Handler{service: service}
}
