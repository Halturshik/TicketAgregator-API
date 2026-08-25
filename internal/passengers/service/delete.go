package service

import (
	"context"
	"errors"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/passengers"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
)

func (s *Service) Delete(ctx context.Context, ownerUserID int, passengerID int) error {
	if passengerID <= 0 {
		return apierror.ErrInvalidRequest
	}
	err := s.repo.Delete(ctx, ownerUserID, passengerID)
	if errors.Is(err, passengers.ErrNotFound) {
		return apierror.ErrNotFound
	}
	if err != nil {
		logger.Error("Ошибка удаления сохраненного пассажира userID=%d passengerID=%d: %v", ownerUserID, passengerID, err)
		return err
	}
	logger.Info("Сохраненный пассажир удален: userID=%d passengerID=%d", ownerUserID, passengerID)
	return nil
}
