package handlers

import "github.com/Halturshik/TicketAgregator-API/internal/bookingaccess"

type Handler struct {
	service bookingaccess.Service
}

func New(service bookingaccess.Service) *Handler {
	return &Handler{service: service}
}
