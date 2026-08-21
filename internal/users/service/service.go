package service

import (
	"github.com/Halturshik/TicketAgregator-API/internal/users"
	"github.com/Halturshik/TicketAgregator-API/internal/users/repository"
)

type Service struct {
	repo *repository.Repository
}

func NewService(repo *repository.Repository) users.Service {
	return &Service{repo: repo}
}
