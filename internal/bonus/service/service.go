package service

import (
	"context"

	"github.com/Halturshik/TicketAgregator-API/internal/bonus"
	"github.com/Halturshik/TicketAgregator-API/internal/bonus/repository"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
)

type Service struct {
	repo *repository.Repository
}

func NewService(repo *repository.Repository) bonus.Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, userID int, limit int, offset int) ([]bonus.Transaction, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	items, err := s.repo.List(ctx, userID, limit, offset)
	if err != nil {
		logger.Error("Ошибка получения истории бонусов userID=%d: %v", userID, err)
		return nil, err
	}
	return items, nil
}
