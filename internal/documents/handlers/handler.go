package handlers

import "github.com/Halturshik/TicketAgregator-API/internal/documents"

type Handler struct {
	service documents.Service
}

func New(service documents.Service) *Handler {
	return &Handler{service: service}
}
