package service

import (
	"context"

	"github.com/Halturshik/TicketAgregator-API/internal/bonus"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
)

func (s *Service) List(ctx context.Context, userID int, limit int, offset int) ([]bonus.Transaction, error) {
	limit = bonus.NormalizeListLimit(limit)
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
