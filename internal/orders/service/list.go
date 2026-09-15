package service

import (
	"context"

	"github.com/Halturshik/TicketAgregator-API/internal/orders"
)

func (s *Service) List(ctx context.Context, userID int, filter orders.ListFilter) (*orders.HistoryPage, error) {
	filter.Limit = orders.NormalizeListLimit(filter.Limit)
	if filter.Offset < 0 {
		filter.Offset = 0
	}
	result, err := s.repo.List(ctx, userID, filter)
	if err != nil {
		return nil, err
	}
	return result, nil
}
