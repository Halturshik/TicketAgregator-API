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

	return nil
}

func (s *ResetPasswordStore) IsVerified(ctx context.Context, email string) (bool, error) {
	exists, err := s.redis.Exists(ctx, s.key(email)).Result()
	if err != nil {
		logger.Error("Ошибка при проверке подтверждения верификации из redis для %s: %v", email, err)
		return false, err
	}

	return exists == 1, nil
}

var consumeResetVerification = redis.NewScript(`
if redis.call("EXISTS", KEYS[1]) == 0 then
    return 0
end
redis.call("DEL", KEYS[1])
return 1
`)

func (s *ResetPasswordStore) ConsumeVerified(ctx context.Context, email string) (bool, error) {
	consumed, err := consumeResetVerification.Run(ctx, s.redis, []string{s.key(email)}).Int()
	if err != nil {
		logger.Error("Ошибка использования подтверждения восстановления пароля для %s: %v", email, err)
		return false, err
	}
	return consumed == 1, nil
}
