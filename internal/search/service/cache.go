package service

import (
	"context"

	"github.com/Halturshik/TicketAgregator-API/internal/search"
)

func (s *Service) cacheSearchResult(ctx context.Context, in search.SearchInput, items []search.TripOption) (*search.CachedResult, error) {
	result := &search.CachedResult{
		SearchID: s.newID(), Input: in, Total: len(items), Items: items,
	}
	if err := s.store.Save(ctx, result); err != nil {
		return nil, err
	}
	return result, nil
}
