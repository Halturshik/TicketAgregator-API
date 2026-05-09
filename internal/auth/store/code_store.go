package store

import (
	"context"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
	"github.com/redis/go-redis/v9"
)

type CodeStore struct {
	redis *redis.Client
}

func NewCodeService(redis *redis.Client) *CodeStore {
	return &CodeStore{redis: redis}
}

func (c *CodeStore) RequestCode(ctx context.Context, email string, code string) error {
	rateKey := RateLimitKey(email)
	if cnt, err := c.redis.Exists(ctx, rateKey).Result(); err != nil {
		return err
	} else if cnt == 1 {
		logger.Warn("Лимит на запрос кода для %s", email)
		return apierror.ErrCodeRateLimited
	}

	cooldownKey := CooldownKey(email)
	if cnt, err := c.redis.Exists(ctx, cooldownKey).Result(); err != nil {
		return err
	} else if cnt == 1 {
		logger.Warn("Запрос кода с блокировкой для %s", email)
		return apierror.ErrTooManyAttempts
	}

	codeKey := CodeKey(email)
	attemptsKey := AttemptsKey(email)

	c.redis.Del(ctx, codeKey, attemptsKey)

	if err := c.redis.Set(ctx, codeKey, code, VerifyCodeTTL).Err(); err != nil {
		return err
	}

	if err := c.redis.Set(ctx, rateKey, "1", CodeRateLimitWindow).Err(); err != nil {
		return err
	}

	return nil
}

func (c *CodeStore) Verify(ctx context.Context, email, code string) error {
	attemptsKey := AttemptsKey(email)
	cooldownKey := CooldownKey(email)
	codeKey := CodeKey(email)
	lockKey := "lock:verify:" + codeKey

	if cnt, err := c.redis.Exists(ctx, cooldownKey).Result(); err != nil {
		return err
	} else if cnt == 1 {
		logger.Warn("Попытка ввода кода во время активной блокировки для %s", email)
		return apierror.ErrTooManyAttempts
	}

	ok, err := c.redis.SetNX(ctx, lockKey, "1", 5*time.Second).Result()
	if err != nil {
		return err
	}
	if !ok {
		logger.Warn("Попытка паралелльного ввода кода для %s", email)
		return apierror.ErrInvalidVerificationCode
	}

	defer c.redis.Del(ctx, lockKey)

	storedCode, err := c.redis.Get(ctx, codeKey).Result()
	if err != nil {
		if err == redis.Nil {
			logger.Warn("Попытка ввода истекшего кода для %s", email)
			return apierror.ErrCodeExpired
		}
		return err
	}

	if storedCode != code {
		attempts, err := c.redis.Incr(ctx, attemptsKey).Result()
		if err != nil {
			return err
		}

		if attempts == 1 {
			if err := c.redis.Expire(ctx, attemptsKey, VerifyCodeTTL).Err(); err != nil {
				return err
			}
		}

		if attempts >= MaxVerifyAttempts {
			logger.Warn("Превышен лимит попыток ввода кода для %s (попыток: %d)", email, attempts)
			if err := c.redis.Set(ctx, cooldownKey, "1", CooldownAfterFailed).Err(); err != nil {
				return err
			}
			c.redis.Del(ctx, codeKey, attemptsKey)

			return apierror.ErrTooManyAttempts
		}

		logger.Warn("Неверный код подтверждения для %s (попытка %d/%d)", email, attempts, MaxVerifyAttempts)
		return apierror.ErrInvalidVerificationCode
	}

	if err := c.redis.Del(ctx, codeKey, attemptsKey).Err(); err != nil {
		logger.Warn("Ошибка при инвалидации кода для %s: %v", email, err)
		return err
	}

	logger.Info("Код и попытки успешно очищены для %s", email)

	return nil
}
