package handlers

import "github.com/Halturshik/TicketAgregator-API/internal/orders"

type Handler struct {
	service orders.Service
}

func New(service orders.Service) *Handler {
	return &Handler{service: service}
}
