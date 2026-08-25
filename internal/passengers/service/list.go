package service

import (
	"context"

	"github.com/Halturshik/TicketAgregator-API/internal/passengers"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
)

func (s *Service) List(ctx context.Context, ownerUserID int) ([]passengers.Passenger, error) {
	items, err := s.repo.List(ctx, ownerUserID)
	if err != nil {
		logger.Error("Ошибка получения сохраненных пассажиров userID=%d: %v", ownerUserID, err)
		return nil, err
	}
	return items, nil
}
