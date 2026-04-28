package store

import (
	"context"

	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
	"github.com/redis/go-redis/v9"
)

type ResetPasswordStore struct {
	redis *redis.Client
}

func NewResetStore(redis *redis.Client) *ResetPasswordStore {
	return &ResetPasswordStore{
		redis: redis,
	}
}

func (s *ResetPasswordStore) key(email string) string {
	return ResetVerifiedKey(email)
}

func (s *ResetPasswordStore) SaveVerified(ctx context.Context, email string) error {
	if err := s.redis.Set(ctx, s.key(email), "true", ResetPasswordTTL).Err(); err != nil {
		logger.Error("Ошибка при сохранении в redis подтверждения верификации для %s: %v", email, err)
		return err
	}

	logger.Info("Подтверждение верификации сохранено в redis для %s (TTL: %v)", email, ResetPasswordTTL)
	return nil
}

func (s *ResetPasswordStore) IsVerified(ctx context.Context, email string) (bool, error) {
	exists, err := s.redis.Exists(ctx, s.key(email)).Result()
	if err != nil {
		logger.Error("Ошибка при проверке подтверждения верификации из redis для %s: %v", email, err)
		return false, err
	}

	logger.Info("Подтверждение верификации получено из redis для %s", email)
	return exists == 1, nil
}

func (s *ResetPasswordStore) Delete(ctx context.Context, email string) error {
	if err := s.redis.Del(ctx, s.key(email)).Err(); err != nil {
		logger.Error("Ошибка при удалении подтверждения верификации из redis для %s: %v", email, err)
		return err
	}

	logger.Info("Подтверждение верификации удалено из redis для %s", email)
	return nil
}
