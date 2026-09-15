package service

import (
	"context"
	"errors"
	"log/slog"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/passengers"
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
		return err
	}
	slog.InfoContext(ctx, "Сохранённый пассажир удалён",
		slog.Int("user_id", ownerUserID),
		slog.Int("passenger_id", passengerID),
	)
	return nil
}
