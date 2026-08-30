package handlers

import "github.com/Halturshik/TicketAgregator-API/internal/refunds"

type Handler struct {
	service refunds.Service
}

func New(service refunds.Service) *Handler {
	return &Handler{service: service}
}
