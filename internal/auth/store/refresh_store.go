package store

import (
	"context"

	"github.com/Halturshik/TicketAgregator-API/internal/auth/token"
	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
	"github.com/redis/go-redis/v9"
)

type RefreshStore struct {
	redis *redis.Client
}

func NewRefreshStore(redis *redis.Client) *RefreshStore {
	return &RefreshStore{redis: redis}
}

func (s *RefreshStore) Save(ctx context.Context, userID int64, refreshToken string) error {
	hash := token.HashToken(refreshToken)
	key := RefreshKey(hash)
	if err := s.redis.Set(ctx, key, userID, token.RefreshTokenTTL).Err(); err != nil {
		logger.Error("Ошибка при сохранении в redis refresh-токена для userID: %v: %v", userID, err)
		return err
	}

	logger.Info("Refresh-токен сохранен в redis для userID: %v (TTL: %v)", userID, token.RefreshTokenTTL)
	return nil
}

func (s *RefreshStore) Get(ctx context.Context, refreshToken string) (int64, error) {
	hash := token.HashToken(refreshToken)
	key := RefreshKey(hash)
	val, err := s.redis.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, apierror.ErrInvalidToken
	}

	if err != nil {
		return 0, err
	}

	logger.Info("Refresh-токен получен из redis: %s...", refreshToken[:8])
	return val, nil
}

func (s *RefreshStore) Delete(ctx context.Context, refreshToken string) error {
	hash := token.HashToken(refreshToken)
	key := RefreshKey(hash)
	if err := s.redis.Del(ctx, key).Err(); err != nil {
		return err
	}

	logger.Info("Refresh-токен удален из redis: %s...", refreshToken[:8])
	return nil
}

func (s *RefreshStore) AddToUserSet(ctx context.Context, userID int64, tokenHash string) error {
	key := RefreshUserSetKey(userID)
	if err := s.redis.SAdd(ctx, key, tokenHash).Err(); err != nil {
		logger.Error("Ошибка при сохранении refresh-токена в список redis по userID: %d: %v", userID, err)
		return err
	}

	s.redis.Expire(ctx, key, token.RefreshTokenTTL*2)
	logger.Info("Refresh-токен сохранен в список redis по userID: %v", userID)
	return nil
}

func (s *RefreshStore) RemoveFromUserSet(ctx context.Context, userID int64, tokenHash string) error {
	key := RefreshUserSetKey(userID)
	removed, err := s.redis.SRem(ctx, key, tokenHash).Result()
	if err != nil {
		logger.Error("Ошибка при удалении refresh-токена из списка в redis по userID: %d: %v", userID, err)
		return err
	}
	if removed == 0 {
		logger.Warn("Refresh-токен не найден из списка в redis при удалении по userID: %d", userID)
	} else {
		logger.Info("Refresh-токен успешно удалён из списка в redis по userID %d", userID)
	}
	return nil
}

func (s *RefreshStore) DeleteAllForUser(ctx context.Context, userID int64) error {
	setKey := RefreshUserSetKey(userID)

	hashes, err := s.redis.SMembers(ctx, setKey).Result()
	if err != nil && err != redis.Nil {
		logger.Error("Ошибка при получении всех хэшей refresh-токенов из списка в redis по userID %d: %v", userID, err)
		return err
	}

	for _, hash := range hashes {
		tokenKey := RefreshKey(hash)
		if err := s.redis.Del(ctx, tokenKey).Err(); err != nil {
			logger.Error("Ошибка при удалении refresh-ключа (%s) из redis по userID %d: %v", hash[:8], userID, err)
		}
	}

	if err := s.redis.Del(ctx, setKey).Err(); err != nil {
		logger.Error("Ошибка при удалении списка активных refresh-токенов в redis по userID %d: %v", userID, err)
		return err
	}

	logger.Info("Все refresh-токены userID (%d) удалены (Set + %d ключей)", userID, len(hashes))
	return nil
}
