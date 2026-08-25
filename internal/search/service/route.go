package service

import (
	"context"
	"errors"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
	"github.com/Halturshik/TicketAgregator-API/internal/search"
	transportpkg "github.com/Halturshik/TicketAgregator-API/internal/transport"
)

type searchRoute struct {
	from   search.City
	to     search.City
	cities []search.City
}

func (s *Service) loadRoute(ctx context.Context, in search.SearchInput) (searchRoute, error) {
	from, err := s.repo.GetCity(ctx, in.FromCityID)
	if errors.Is(err, search.ErrCityNotFound) {
		return searchRoute{}, apierror.ErrNotFound
	}
	if err != nil {
		logger.Error("Ошибка получения города отправления: %v", err)
		return searchRoute{}, err
	}
	to, err := s.repo.GetCity(ctx, in.ToCityID)
	if errors.Is(err, search.ErrCityNotFound) {
		return searchRoute{}, apierror.ErrNotFound
	}
	if err != nil {
		logger.Error("Ошибка получения города прибытия: %v", err)
		return searchRoute{}, err
	}
	if from.ID == to.ID {
		return searchRoute{}, apierror.ErrInvalidRequest
	}
	cities, err := s.repo.ListCities(ctx)
	if err != nil {
		logger.Error("Ошибка получения списка городов для генерации поиска: %v", err)
		return searchRoute{}, err
	}
	return searchRoute{from: *from, to: *to, cities: cities}, nil
}

func validateTransportRoute(transport string, from search.City, to search.City) error {
	if transportpkg.IsGround(transport) && !groundRouteAllowed(from, to) {
		return apierror.ErrInvalidRequest
	}
	return nil
}

func groundRouteAllowed(from search.City, to search.City) bool {
	if from.CountryID == to.CountryID {
		return true
	}
	for _, neighborID := range from.NeighborCountryIDs {
		if neighborID == to.CountryID {
			return true
		}
	}
	return false
}
