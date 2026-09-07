package service

import "github.com/Halturshik/TicketAgregator-API/internal/trips"

type Service struct {
	repo trips.Repository
}

var _ trips.Service = (*Service)(nil)

func NewService(repo trips.Repository) trips.Service {
	return &Service{repo: repo}
}
