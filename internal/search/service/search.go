package service

import (
	"context"

	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
	"github.com/Halturshik/TicketAgregator-API/internal/search"
)

func (s *Service) Search(ctx context.Context, transport string, in search.SearchInput, userID *int) (*search.SearchResult, error) {
	now := s.now()
	searchDate, err := normalizeInput(&in, now)
	if err != nil {
		return nil, err
	}
	provider, err := s.provider(transport)
	if err != nil {
		return nil, err
	}
	route, err := s.loadRoute(ctx, in)
	if err != nil {
		return nil, err
	}
	if err := validateTransportRoute(transport, route.from, route.to); err != nil {
		return nil, err
	}

	total := generatedOffersCount(searchDate, dateOnly(now), now.UnixNano())
	items, err := generateTripOptions(ctx, provider, in, route, total)
	if err != nil {
		logger.Error("Ошибка генерации mock-предложений: %v", err)
		return nil, err
	}
	cached, err := s.cacheSearchResult(ctx, in, items)
	if err != nil {
		return nil, err
	}

	logger.Info("Сгенерирован результат поиска searchID=%s total=%d transport=%s", cached.SearchID, cached.Total, transport)
	return s.page(cached, in.Offset, in.Limit, userID), nil
}
