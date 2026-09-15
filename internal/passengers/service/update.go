package service

import (
	"context"
	"errors"
	"log/slog"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/passengers"
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
		return nil, err
	}
	slog.InfoContext(ctx, "Сохранённый пассажир обновлён",
		slog.Int("user_id", ownerUserID),
		slog.Int("passenger_id", passengerID),
	)
	return item, nil
}
