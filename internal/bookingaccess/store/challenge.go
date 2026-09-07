package store

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Halturshik/TicketAgregator-API/internal/bookingaccess"
	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
	"github.com/redis/go-redis/v9"
)

type challengePayload struct {
	Challenge bookingaccess.Challenge `json:"challenge"`
	Code      string                  `json:"code"`
	Identity  string                  `json:"identity"`
}

type ChallengeStore struct {
	redis *redis.Client
}

func NewChallengeStore(redisClient *redis.Client) *ChallengeStore {
	return &ChallengeStore{redis: redisClient}
}

func (s *ChallengeStore) Request(ctx context.Context, challenge bookingaccess.Challenge, code string) error {
	identity := challengeIdentity(challenge.Locator, challenge.Email)
	blocked, err := s.redis.Exists(ctx, cooldownKey(identity)).Result()
	if err != nil {
		return err
	}
	if blocked == 1 {
		logger.Warn("Запрос доступа к заказу во время блокировки order_id=%d", challenge.OrderID)
		return apierror.ErrTooManyAttempts
	}

	claimed, err := s.redis.SetNX(ctx, requestKey(identity), "1", bookingaccess.CodeRequestInterval).Result()
	if err != nil {
		return err
	}
	if !claimed {
		logger.Warn("Превышен лимит запроса кода доступа к заказу order_id=%d", challenge.OrderID)
		return apierror.ErrCodeRateLimited
	}
	keepLimit := false
	defer func() {
		if !keepLimit {
			s.redis.Del(ctx, requestKey(identity))
		}
	}()

	payload, err := json.Marshal(challengePayload{Challenge: challenge, Code: code, Identity: identity})
	if err != nil {
		return fmt.Errorf("marshal booking access challenge: %w", err)
	}
	previousID, err := s.redis.Get(ctx, activeChallengeKey(identity)).Result()
	if err != nil && err != redis.Nil {
		return err
	}
	transaction := s.redis.TxPipeline()
	if previousID != "" {
		transaction.Del(ctx, challengeKey(previousID), attemptsKey(previousID))
	}
	transaction.Set(ctx, challengeKey(challenge.ID), payload, bookingaccess.VerificationCodeTTL)
	transaction.Set(ctx, activeChallengeKey(identity), challenge.ID, bookingaccess.VerificationCodeTTL)
	if _, err := transaction.Exec(ctx); err != nil {
		return err
	}
	keepLimit = true
	return nil
}

func (s *ChallengeStore) Verify(ctx context.Context, challengeID string, code string) (*bookingaccess.Challenge, error) {
	locked, err := s.redis.SetNX(
		ctx, verifyLockKey(challengeID), "1", bookingaccess.VerificationLockTTL,
	).Result()
	if err != nil {
		return nil, err
	}
	if !locked {
		return nil, apierror.ErrInvalidVerificationCode
	}
	defer s.redis.Del(ctx, verifyLockKey(challengeID))

	value, err := s.redis.Get(ctx, challengeKey(challengeID)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, apierror.ErrCodeExpired
		}
		return nil, err
	}
	var payload challengePayload
	if err := json.Unmarshal(value, &payload); err != nil {
		return nil, fmt.Errorf("decode booking access challenge: %w", err)
	}
	blocked, err := s.redis.Exists(ctx, cooldownKey(payload.Identity)).Result()
	if err != nil {
		return nil, err
	}
	if blocked == 1 {
		return nil, apierror.ErrTooManyAttempts
	}
	if payload.Code != code {
		return nil, s.recordFailedAttempt(ctx, challengeID, payload)
	}
	if err := s.redis.Del(
		ctx, challengeKey(challengeID), attemptsKey(challengeID), activeChallengeKey(payload.Identity),
	).Err(); err != nil {
		return nil, err
	}
	return &payload.Challenge, nil
}

func (s *ChallengeStore) recordFailedAttempt(
	ctx context.Context,
	challengeID string,
	payload challengePayload,
) error {
	attempts, err := s.redis.Incr(ctx, attemptsKey(challengeID)).Result()
	if err != nil {
		return err
	}
	if attempts == 1 {
		if err := s.redis.Expire(ctx, attemptsKey(challengeID), bookingaccess.VerificationCodeTTL).Err(); err != nil {
			return err
		}
	}
	if attempts >= bookingaccess.MaxVerificationAttempts {
		if err := s.redis.Set(
			ctx,
			cooldownKey(payload.Identity),
			"1",
			bookingaccess.VerificationCooldown,
		).Err(); err != nil {
			return err
		}
		s.redis.Del(
			ctx, challengeKey(challengeID), attemptsKey(challengeID), activeChallengeKey(payload.Identity),
		)
		logger.Warn("Превышен лимит проверки доступа к заказу order_id=%d", payload.Challenge.OrderID)
		return apierror.ErrTooManyAttempts
	}
	logger.Warn(
		"Неверный код доступа к заказу order_id=%d попытка=%d/%d",
		payload.Challenge.OrderID,
		attempts,
		bookingaccess.MaxVerificationAttempts,
	)
	return apierror.ErrInvalidVerificationCode
}

func challengeIdentity(locator string, email string) string {
	value := strings.ToUpper(strings.TrimSpace(locator)) + ":" + strings.ToLower(strings.TrimSpace(email))
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
