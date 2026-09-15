package store

import (
	"context"
	"fmt"
	"strconv"

	"github.com/redis/go-redis/v9"
)

type LoginStore struct {
	redis *redis.Client
}

func NewLoginStore(redis *redis.Client) *LoginStore {
	return &LoginStore{
		redis: redis,
	}
}

func (s *LoginStore) key(email string) string {
	return LoginKey(email)
}

func (s *LoginStore) Save(ctx context.Context, email string, userID int64) error {
	if err := s.redis.Set(ctx, s.key(email), userID, LoginTTL).Err(); err != nil {
		return fmt.Errorf("save login data: %w", err)
	}

	return nil
}

func (s *LoginStore) Get(ctx context.Context, email string) (int64, error) {
	val, err := s.redis.Get(ctx, s.key(email)).Result()
	if err != nil {
		return 0, fmt.Errorf("get login data: %w", err)
	}

	userID, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse login user id: %w", err)
	}
	return userID, nil
}

func (s *LoginStore) Delete(ctx context.Context, email string) error {
	if err := s.redis.Del(ctx, s.key(email)).Err(); err != nil {
		return fmt.Errorf("delete login data: %w", err)
	}

	return nil
}
