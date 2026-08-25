package service

import (
	"context"

	"github.com/Halturshik/TicketAgregator-API/internal/documents"
)

func (s *Service) List(ctx context.Context, ownerUserID int, passengerID *int) ([]documents.Document, error) {
	if passengerID != nil {
		return s.repo.ListForPassenger(ctx, ownerUserID, *passengerID)
	}
	return s.repo.ListForUser(ctx, ownerUserID)
}
