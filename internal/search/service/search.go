package service

import (
	"context"
	"database/sql"
	"sort"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
	"github.com/Halturshik/TicketAgregator-API/internal/search"
	"github.com/Halturshik/TicketAgregator-API/internal/search/provider"
	"github.com/google/uuid"
)

func (s *Service) Search(ctx context.Context, transport string, in search.SearchInput, userID *int) (*search.SearchResult, error) {
	if err := normalizeInput(&in); err != nil {
		return nil, err
	}
	if _, ok := s.providers[transport]; !ok {
		return nil, apierror.ErrInvalidRequest
	}

	from, to, cities, err := s.loadRoute(ctx, in)
	if err != nil {
		return nil, err
	}

	total := MinGeneratedOffers + int(time.Now().UnixNano()%int64(MaxExtraOffers))
	if (transport == "rail" || transport == "bus") && !groundRouteAllowed(*from, *to) {
		return nil, apierror.ErrInvalidRequest
	}
	items, err := s.generateOffers(ctx, transport, in, *from, *to, cities, total)
	if err != nil {
		logger.Error("Ошибка генерации mock-предложений: %v", err)
		return nil, err
	}

	searchID := uuid.NewString()
	cached := &search.CachedResult{SearchID: searchID, Input: in, Total: len(items), Items: items}
	if err := s.store.Save(ctx, cached); err != nil {
		logger.Error("Ошибка кеширования результата поиска searchID=%s: %v", searchID, err)
		return nil, err
	}

	logger.Info("Сгенерирован результат поиска searchID=%s total=%d transport=%s", searchID, len(items), transport)
	return s.page(cached, in.Offset, in.Limit, userID), nil
}

func (s *Service) loadRoute(ctx context.Context, in search.SearchInput) (*search.City, *search.City, []search.City, error) {
	from, err := s.repo.GetCity(ctx, in.FromCityID)
	if err == sql.ErrNoRows {
		return nil, nil, nil, apierror.ErrNotFound
	}
	if err != nil {
		logger.Error("Ошибка получения города отправления: %v", err)
		return nil, nil, nil, err
	}
	to, err := s.repo.GetCity(ctx, in.ToCityID)
	if err == sql.ErrNoRows {
		return nil, nil, nil, apierror.ErrNotFound
	}
	if err != nil {
		logger.Error("Ошибка получения города прибытия: %v", err)
		return nil, nil, nil, err
	}
	if from.ID == to.ID {
		return nil, nil, nil, apierror.ErrInvalidRequest
	}
	cities, err := s.repo.ListCities(ctx)
	if err != nil {
		logger.Error("Ошибка получения списка городов для генерации поиска: %v", err)
		return nil, nil, nil, err
	}
	return from, to, cities, nil
}

func (s *Service) generateOffers(ctx context.Context, transport string, in search.SearchInput, from search.City, to search.City, cities []search.City, total int) ([]search.Offer, error) {
	offers, err := s.providers[transport].Generate(ctx, provider.Request{
		Input:  in,
		From:   from,
		To:     to,
		Cities: cities,
		Count:  total,
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(offers, func(i, j int) bool {
		return offers[i].Price < offers[j].Price
	})
	return offers, nil
}

func normalizeInput(in *search.SearchInput) error {
	if in.Passengers <= 0 {
		in.Passengers = 1
	}
	if in.Limit <= 0 || in.Limit > MaxPageSize {
		in.Limit = DefaultPageSize
	}
	if _, err := time.Parse("2006-01-02", in.Date); err != nil {
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
