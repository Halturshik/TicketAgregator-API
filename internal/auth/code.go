package auth

import (
	"context"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/logger"
	"github.com/redis/go-redis/v9"
)

const (
	emailVerify             = "email_verify:"
	emailVerifyAttempts     = "email_verify_attempts:"
	codeRequestRateLimit    = "code_request_rate_limit:"
	codeCooldownAfterFailed = "code_cooldown_after_failed:"

	maxVerifyAttempts   = 5
	verifyCodeTTL       = 5 * time.Minute
	codeRateLimitWindow = 2 * time.Minute
	cooldownAfterFailed = 5 * time.Minute
)

type CodeService struct {
	redis *redis.Client
}

func NewCodeService(redis *redis.Client) *CodeService {
	return &CodeService{redis: redis}
}

func (c *CodeService) Generate(ctx context.Context, email string) (string, error) {
	rateKey := codeRequestRateLimit + email
	if cnt, err := c.redis.Exists(ctx, rateKey).Result(); err != nil {
		return "", err
	} else if cnt == 1 {
		logger.Warn("Лимит на запрос кода для %s", email)
		return "", apierror.ErrCodeRateLimited
	}

	cooldownKey := codeCooldownAfterFailed + email
	if cnt, err := c.redis.Exists(ctx, cooldownKey).Result(); err != nil {
		return "", err
	} else if cnt == 1 {
		logger.Warn("Запрос кода с блокировкой для %s", email)
		return "", apierror.ErrTooManyAttempts
	}

	codeKey := emailVerify + email
	attemptsKey := emailVerifyAttempts + email

	c.redis.Del(ctx, codeKey)
	c.redis.Del(ctx, attemptsKey)

	code, err := GenerateVerificationCode()
	if err != nil {
		logger.Error("Ошибка при генерации кода для %s: %v", email, err)
		return "", err
	}

	if err := c.redis.Set(ctx, codeKey, code, verifyCodeTTL).Err(); err != nil {
		return "", err
	}

	if err := c.redis.Set(ctx, rateKey, "1", codeRateLimitWindow).Err(); err != nil {
		return "", err
	}

	logger.Info("Код подтверждения сгенерирован для %s", email)
	return code, nil
}

func (c *CodeService) Verify(ctx context.Context, email, code string) error {
	attemptsKey := emailVerifyAttempts + email
	cooldownKey := codeCooldownAfterFailed + email
	codeKey := emailVerify + email

	if cnt, err := c.redis.Exists(ctx, cooldownKey).Result(); err != nil {
		return err
	} else if cnt == 1 {
		logger.Warn("Попытка ввода кода во время активной блокировки для %s", email)
		return apierror.ErrTooManyAttempts
	}

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
			if err := c.redis.Expire(ctx, attemptsKey, verifyCodeTTL).Err(); err != nil {
				return err
			}
		}

		if attempts >= maxVerifyAttempts {
			logger.Warn("Превышен лимит попыток ввода кода для %s (попыток: %d)", email, attempts)
			if err := c.redis.Set(ctx, cooldownKey, "1", cooldownAfterFailed).Err(); err != nil {
				return err
			}
			c.redis.Del(ctx, codeKey)
			c.redis.Del(ctx, attemptsKey)

			return apierror.ErrTooManyAttempts
		}

		logger.Warn("Неверный код подтверждения для %s (попытка %d/%d)", email, attempts, maxVerifyAttempts)
		return apierror.ErrInvalidVerificationCode
	}

	logger.Info("Код подтверждения успешно введен для %s", email)
	return nil
}

func (c *CodeService) Clear(ctx context.Context, email string) error {
	codeKey := emailVerify + email
	attemptsKey := emailVerifyAttempts + email

	if err := c.redis.Del(ctx, codeKey, attemptsKey).Err(); err != nil {
		return err
	}

	logger.Info("Код и попытки успешно очищены для %s", email)
	return nil
}
