package store

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Halturshik/TicketAgregator-API/internal/auth"
	"github.com/redis/go-redis/v9"
)

type RegistrationStore struct {
	redis *redis.Client
}

func NewRegistrationStore(redis *redis.Client) *RegistrationStore {
	return &RegistrationStore{
		redis: redis,
	}
}

func (s *RegistrationStore) key(email string) string {
	return RegistrationKey(email)
}

func (s *RegistrationStore) Save(ctx context.Context, email string, data auth.PendingRegistration) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal pending registration: %w", err)
	}

	if err := s.redis.Set(ctx, s.key(email), bytes, RegistrationTTL).Err(); err != nil {
		return fmt.Errorf("save pending registration: %w", err)
	}

	return nil
}

func (s *RegistrationStore) Get(ctx context.Context, email string) (*auth.PendingRegistration, error) {
	data, err := s.redis.Get(ctx, s.key(email)).Bytes()
	if err != nil {
		return nil, fmt.Errorf("get pending registration: %w", err)
	}

	var result auth.PendingRegistration
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("unmarshal pending registration: %w", err)
	}

	return &result, nil
}

func (s *RegistrationStore) Delete(ctx context.Context, email string) error {
	if err := s.redis.Del(ctx, s.key(email)).Err(); err != nil {
		return fmt.Errorf("delete pending registration: %w", err)
	}

	return nil
}
