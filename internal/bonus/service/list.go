package service

import (
	"context"

	"github.com/Halturshik/TicketAgregator-API/internal/bonus"
)

func (s *Service) List(ctx context.Context, userID int, limit int, offset int) ([]bonus.Transaction, error) {
	limit = bonus.NormalizeListLimit(limit)
	if offset < 0 {
		offset = 0
	}
	items, err := s.repo.List(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	return items, nil
}
