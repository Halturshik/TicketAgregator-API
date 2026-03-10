package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/auth/types"
	"github.com/Halturshik/TicketAgregator-API/internal/logger"
	"github.com/redis/go-redis/v9"
)

type RegistrationStore struct {
	redis *redis.Client
	ttl   time.Duration
}

func NewRegistrationStore(redis *redis.Client, ttl time.Duration) *RegistrationStore {
	return &RegistrationStore{
		redis: redis,
		ttl:   RegistrationTTL,
	}
}

func (s *RegistrationStore) key(email string) string {
	return RegistrationKey(email)
}

func (s *RegistrationStore) Save(ctx context.Context, email string, data types.RegisterInput) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		logger.Error("Ошибка при сериализации регистрационных данных для %s: %v", email, err)
		return err
	}

	if err := s.redis.Set(ctx, s.key(email), bytes, s.ttl).Err(); err != nil {
		logger.Error("Ошибка при сохранении в redis регистрационных данных для %s: %v", email, err)
		return err
	}

	logger.Info("Регистрационные данные сохранены в redis для %s (TTL: %v)", email, s.ttl)
	return nil
}

func (s *RegistrationStore) Get(ctx context.Context, email string) (*types.RegisterInput, error) {
	data, err := s.redis.Get(ctx, s.key(email)).Bytes()
	if err != nil {
		logger.Error("Ошибка при получения регистрационных данных из redis для %s: %v", email, err)
		return nil, err
	}

	var result types.RegisterInput
	if err := json.Unmarshal(data, &result); err != nil {
		logger.Error("Ошибка при десериализации регистрационных данных для %s: %v", email, err)
		return nil, err
	}

	logger.Info("Регистрационные данные получены из redis для %s", email)
	return &result, nil
}

func (s *RegistrationStore) Delete(ctx context.Context, email string) error {
	if err := s.redis.Del(ctx, s.key(email)).Err(); err != nil {
		logger.Error("Ошибка при удалении регистрационных данных из redis для %s: %v", email, err)
		return err
	}

	logger.Info("Регистрационные данные удалены из redis для %s", email)
	return nil
}
