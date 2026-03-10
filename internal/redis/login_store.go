package redis

import (
	"context"
	"strconv"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/logger"
	"github.com/redis/go-redis/v9"
)

type LoginStore struct {
	redis *redis.Client
	ttl   time.Duration
}

func NewLoginStore(redis *redis.Client, ttl time.Duration) *LoginStore {
	return &LoginStore{
		redis: redis,
		ttl:   LoginTTL,
	}
}

func (s *LoginStore) key(email string) string {
	return LoginKey(email)
}

func (s *LoginStore) Save(ctx context.Context, email string, userID int64) error {
	if err := s.redis.Set(ctx, s.key(email), userID, s.ttl).Err(); err != nil {
		logger.Error("Ошибка при сохранении в redis авторизационных данных для %s: %v", email, err)
		return err
	}

	logger.Info("Авторизационные данные сохранены в redis для %s (TTL: %v)", email, s.ttl)
	return nil
}

func (s *LoginStore) Get(ctx context.Context, email string) (int64, error) {
	val, err := s.redis.Get(ctx, s.key(email)).Result()
	if err != nil {
		logger.Error("Ошибка при получения авторизационных данных из redis для %s: %v", email, err)
		return 0, err
	}

	logger.Info("Авторизационные данные получены из redis для %s", email)
	return strconv.ParseInt(val, 10, 64)
}

func (s *LoginStore) Delete(ctx context.Context, email string) error {
	if err := s.redis.Del(ctx, s.key(email)).Err(); err != nil {
		logger.Error("Ошибка при удалении авторизационных данных из redis для %s: %v", email, err)
		return err
	}

	logger.Info("Авторизационные данные удалены из redis для %s", email)
	return nil
}
