package service

import "github.com/Halturshik/TicketAgregator-API/internal/passengers"

type Service struct {
	repo passengers.Repository
}

func NewService(repo passengers.Repository) passengers.Service {
	return &Service{repo: repo}
}
