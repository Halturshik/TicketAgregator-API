package service

import "github.com/Halturshik/TicketAgregator-API/internal/users"

type Service struct {
	repo users.Repository
}

func NewService(repo users.Repository) users.Service {
	return &Service{repo: repo}
}
