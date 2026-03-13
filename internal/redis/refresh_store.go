package redis

import (
	"context"

	"github.com/Halturshik/TicketAgregator-API/internal/authutils"
	"github.com/Halturshik/TicketAgregator-API/internal/logger"
	"github.com/redis/go-redis/v9"
)

type RefreshStore struct {
	redis *redis.Client
}

func NewRefreshStore(redis *redis.Client) *RefreshStore {
	return &RefreshStore{redis: redis}
}

func (s *RefreshStore) Save(ctx context.Context, userID int64, token string) error {
	hash := authutils.HashToken(token)
	key := RefreshKey(hash)
	if err := s.redis.Set(ctx, key, userID, authutils.RefreshTokenTTL).Err(); err != nil {
		logger.Error("Ошибка при сохранении в redis refresh-токена данных для userID: %v: %v", userID, err)
		return err
	}

	logger.Info("Refresh-токен сохранен в redis для userID: %v (TTL: %v)", userID, authutils.RefreshTokenTTL)
	return nil
}

func (s *RefreshStore) Get(ctx context.Context, token string) (int64, error) {
	hash := authutils.HashToken(token)
	key := RefreshKey(hash)
	val, err := s.redis.Get(ctx, key).Int64()
	if err != nil {
		return 0, err
	}

	logger.Info("Refresh-токен получен из redis: %s...", token[:8])
	return val, nil
}

func (s *RefreshStore) Delete(ctx context.Context, token string) error {
	hash := authutils.HashToken(token)
	key := RefreshKey(hash)
	if err := s.redis.Del(ctx, key).Err(); err != nil {
		return err
	}

	logger.Info("Refresh-токен удален из redis: %s...", token[:8])
	return nil
}
