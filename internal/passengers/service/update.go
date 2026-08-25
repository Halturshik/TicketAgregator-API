package service

import (
	"context"
	"errors"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/passengers"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
)

func (s *Service) Update(ctx context.Context, ownerUserID int, passengerID int, in passengers.SavePassengerInput) (*passengers.Passenger, error) {
	if passengerID <= 0 {
		return nil, apierror.ErrInvalidRequest
	}
	birthDate, err := s.validate(&in)
	if err != nil {
		return nil, err
	}
	item, err := s.repo.Update(ctx, ownerUserID, passengerID, in, birthDate)
	if errors.Is(err, passengers.ErrNotFound) {
		return nil, apierror.ErrNotFound
	}
	if err != nil {
		logger.Error("Ошибка обновления сохраненного пассажира userID=%d passengerID=%d: %v", ownerUserID, passengerID, err)
		return nil, err
	}
	logger.Info("Сохраненный пассажир обновлен: userID=%d passengerID=%d", ownerUserID, passengerID)
	return item, nil
}
