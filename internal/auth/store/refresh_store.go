package store

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Halturshik/TicketAgregator-API/internal/auth/token"
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
		return fmt.Errorf("save refresh token for user %d: %w", userID, err)
	}

	return nil
}

func (s *RefreshStore) Delete(ctx context.Context, refreshToken string) error {
	hash := token.HashToken(refreshToken)
	key := RefreshKey(hash)
	if err := s.redis.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("delete refresh token: %w", err)
	}

	return nil
}

func (s *RefreshStore) AddToUserSet(ctx context.Context, userID int64, tokenHash string) error {
	key := RefreshUserSetKey(userID)
	if err := s.redis.SAdd(ctx, key, tokenHash).Err(); err != nil {
		return fmt.Errorf("add refresh token to user %d set: %w", userID, err)
	}

	if err := s.redis.Expire(ctx, key, token.RefreshTokenTTL*2).Err(); err != nil {
		return fmt.Errorf("set refresh token set ttl for user %d: %w", userID, err)
	}
	slog.DebugContext(ctx, "Refresh-токен добавлен в список пользователя", slog.Int64("user_id", userID))
	return nil
}

func (s *RefreshStore) RemoveFromUserSet(ctx context.Context, userID int64, tokenHash string) error {
	key := RefreshUserSetKey(userID)
	removed, err := s.redis.SRem(ctx, key, tokenHash).Result()
	if err != nil {
		return fmt.Errorf("remove refresh token from user %d set: %w", userID, err)
	}
	if removed == 0 {
		slog.WarnContext(ctx, "Refresh-токен не найден в списке пользователя", slog.Int64("user_id", userID))
	} else {
		slog.DebugContext(ctx, "Refresh-токен удалён из списка пользователя", slog.Int64("user_id", userID))
	}
	return nil
}

func (s *RefreshStore) DeleteAllForUser(ctx context.Context, userID int64) error {
	setKey := RefreshUserSetKey(userID)

	hashes, err := s.redis.SMembers(ctx, setKey).Result()
	if err != nil && err != redis.Nil {
		return fmt.Errorf("list refresh tokens for user %d: %w", userID, err)
	}

	for _, hash := range hashes {
		tokenKey := RefreshKey(hash)
		if err := s.redis.Del(ctx, tokenKey).Err(); err != nil {
			slog.ErrorContext(ctx, "Ошибка удаления refresh-токена пользователя",
				slog.Int64("user_id", userID),
				slog.Any("error", err),
			)
		}
	}

	if err := s.redis.Del(ctx, setKey).Err(); err != nil {
		return fmt.Errorf("delete refresh token set for user %d: %w", userID, err)
	}

	slog.InfoContext(ctx, "Все refresh-токены пользователя удалены",
		slog.Int64("user_id", userID),
		slog.Int("token_count", len(hashes)),
	)
	return nil
}
