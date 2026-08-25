package store

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Halturshik/TicketAgregator-API/internal/search"
	"github.com/redis/go-redis/v9"
)

type Store struct {
	redis *redis.Client
}

var _ search.Store = (*Store)(nil)

func New(redis *redis.Client) *Store {
	return &Store{redis: redis}
}

func key(id string) string {
	return "search:result:" + id
}

func (s *Store) Save(ctx context.Context, result *search.CachedResult) error {
	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("marshal search result: %w", err)
	}
	return s.redis.Set(ctx, key(result.SearchID), data, SearchTTL).Err()
}

func (s *Store) Get(ctx context.Context, id string) (*search.CachedResult, error) {
	data, err := s.redis.Get(ctx, key(id)).Bytes()
	if err == redis.Nil {
		return nil, search.ErrCachedResultNotFound
	}
	if err != nil {
		return nil, err
	}
	var result search.CachedResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("unmarshal search result: %w", err)
	}
	return &result, nil
}
