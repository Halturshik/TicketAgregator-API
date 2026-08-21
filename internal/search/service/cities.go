package service

import (
	"context"

	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
	"github.com/Halturshik/TicketAgregator-API/internal/search"
)

func (s *Service) ListCities(ctx context.Context) (*search.CitiesResult, error) {
	items, err := s.repo.ListCitiesPublic(ctx)
	if err != nil {
		logger.Error("Ошибка получения списка городов: %v", err)
		return nil, err
	}
	return &search.CitiesResult{Items: items}, nil
}
