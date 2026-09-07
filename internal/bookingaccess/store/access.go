package store

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"strconv"
	"strings"

	"github.com/Halturshik/TicketAgregator-API/internal/bookingaccess"
	"github.com/redis/go-redis/v9"
)

type AccessStore struct {
	redis *redis.Client
}

const accessTokenBytes = 32

func NewAccessStore(redisClient *redis.Client) *AccessStore {
	return &AccessStore{redis: redisClient}
}

func (s *AccessStore) Issue(ctx context.Context, orderID int) (string, error) {
	raw := make([]byte, accessTokenBytes)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	if err := s.redis.Set(ctx, tokenKey(tokenHash(token)), orderID, bookingaccess.AccessTokenTTL).Err(); err != nil {
		return "", err
	}
	return token, nil
}

func (s *AccessStore) Authorize(ctx context.Context, token string, orderID int) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return bookingaccess.ErrInvalidToken
	}
	stored, err := s.redis.Get(ctx, tokenKey(tokenHash(token))).Result()
	if err != nil {
		if err == redis.Nil {
			return bookingaccess.ErrInvalidToken
		}
		return err
	}
	if stored != strconv.Itoa(orderID) {
		return bookingaccess.ErrInvalidToken
	}
	return nil
}

func tokenHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
