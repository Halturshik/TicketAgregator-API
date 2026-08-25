package service

import (
	"context"
	"errors"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/passengers"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
)

func (s *Service) GetOwned(ctx context.Context, ownerUserID int, passengerID int) (*passengers.Passenger, error) {
	if passengerID <= 0 {
		return nil, apierror.ErrInvalidRequest
	}
	item, err := s.repo.GetOwned(ctx, ownerUserID, passengerID)
	if errors.Is(err, passengers.ErrNotFound) {
		return nil, apierror.ErrNotFound
	}
	if err != nil {
		logger.Error("Ошибка проверки принадлежности пассажира userID=%d passengerID=%d: %v", ownerUserID, passengerID, err)
		return nil, err
	}
	return item, nil
}
