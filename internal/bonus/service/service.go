package service

import "github.com/Halturshik/TicketAgregator-API/internal/bonus"

type Service struct {
	repo bonus.Repository
}

func NewService(repo bonus.Repository) bonus.Service {
	return &Service{repo: repo}
}
