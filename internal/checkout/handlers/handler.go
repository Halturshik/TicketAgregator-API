package handlers

import "github.com/Halturshik/TicketAgregator-API/internal/checkout"

type Handler struct {
	service checkout.Service
}

func New(service checkout.Service) *Handler {
	return &Handler{service: service}
}
