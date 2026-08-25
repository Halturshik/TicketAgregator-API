package service

import (
	"context"
	"sort"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/search"
	"github.com/Halturshik/TicketAgregator-API/internal/search/provider"
)

func (s *Service) provider(transport string) (provider.Provider, error) {
	item, ok := s.providers[transport]
	if !ok {
		return nil, apierror.ErrInvalidRequest
	}
	return item, nil
}

func generateTripOptions(ctx context.Context, item provider.Provider, in search.SearchInput, route searchRoute, total int) ([]search.TripOption, error) {
	options, err := item.Generate(ctx, provider.Request{
		Input: in, From: route.from, To: route.to, Cities: route.cities, Count: total,
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(options, func(i, j int) bool {
		return options[i].Price < options[j].Price
	})
	return options, nil
}
