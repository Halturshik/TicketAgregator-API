package store

import (
	"context"
	"strconv"

	"github.com/Halturshik/TicketAgregator-API/internal/auth/token"
	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
	"github.com/redis/go-redis/v9"
)

var rotateRefreshToken = redis.NewScript(`
local storedUserID = redis.call("GET", KEYS[1])
if not storedUserID or storedUserID ~= ARGV[1] then
    return 0
end

redis.call("DEL", KEYS[1])
redis.call("SREM", KEYS[3], ARGV[2])
redis.call("SET", KEYS[2], ARGV[1], "PX", ARGV[4])
redis.call("SADD", KEYS[3], ARGV[3])
redis.call("PEXPIRE", KEYS[3], ARGV[5])
return 1
`)

func (s *RefreshStore) Rotate(ctx context.Context, userID int64, oldToken string, newToken string) error {
	oldHash := token.HashToken(oldToken)
	newHash := token.HashToken(newToken)
	keys := []string{
		RefreshKey(oldHash),
		RefreshKey(newHash),
		RefreshUserSetKey(userID),
	}

	rotated, err := rotateRefreshToken.Run(ctx, s.redis, keys,
		strconv.FormatInt(userID, 10),
		oldHash,
		newHash,
		strconv.FormatInt(token.RefreshTokenTTL.Milliseconds(), 10),
		strconv.FormatInt((token.RefreshTokenTTL*2).Milliseconds(), 10),
	).Int()
	if err != nil {
		logger.Error("Ошибка атомарной ротации refresh-токена для userID %d: %v", userID, err)
		return err
	}
	if rotated != 1 {
		return apierror.ErrInvalidToken
	}

	return nil
}
