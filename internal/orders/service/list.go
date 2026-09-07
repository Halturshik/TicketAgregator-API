package service

import (
	"context"

	"github.com/Halturshik/TicketAgregator-API/internal/orders"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
)

func (s *Service) List(ctx context.Context, userID int, filter orders.ListFilter) (*orders.HistoryPage, error) {
	filter.Limit = orders.NormalizeListLimit(filter.Limit)
	if filter.Offset < 0 {
		filter.Offset = 0
	}
	result, err := s.repo.List(ctx, userID, filter)
	if err != nil {
		logger.Error("Ошибка получения истории заказов userID=%d: %v", userID, err)
		return nil, err
	}
	return result, nil
}
