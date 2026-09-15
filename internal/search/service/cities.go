package service

import (
	"context"

	"github.com/Halturshik/TicketAgregator-API/internal/search"
)

func (s *Service) ListCities(ctx context.Context) (*search.CitiesResult, error) {
	items, err := s.repo.ListCitiesPublic(ctx)
	if err != nil {
		return nil, err
	}
	return &search.CitiesResult{Items: items}, nil
}
