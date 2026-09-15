package service

import (
	"context"

	"github.com/Halturshik/TicketAgregator-API/internal/passengers"
)

func (s *Service) List(ctx context.Context, ownerUserID int) ([]passengers.Passenger, error) {
	items, err := s.repo.List(ctx, ownerUserID)
	if err != nil {
		return nil, err
	}
	return items, nil
}
