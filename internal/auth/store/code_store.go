package store

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	platformlogger "github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
	"github.com/redis/go-redis/v9"
)

type CodeStore struct {
	redis *redis.Client
}

func NewCodeService(redis *redis.Client) *CodeStore {
	return &CodeStore{redis: redis}
}

func (c *CodeStore) RequestCode(ctx context.Context, email string, code string) error {
	cooldownKey := CooldownKey(email)
	if cnt, err := c.redis.Exists(ctx, cooldownKey).Result(); err != nil {
		return fmt.Errorf("check verification cooldown: %w", err)
	} else if cnt == 1 {
		slog.WarnContext(ctx, "Запрос кода во время блокировки",
			slog.String("email", platformlogger.MaskEmail(email)),
		)
		return apierror.ErrTooManyAttempts
	}

	rateKey := RateLimitKey(email)
	claimed, err := c.redis.SetNX(ctx, rateKey, "1", CodeRateLimitWindow).Result()
	if err != nil {
		return fmt.Errorf("claim verification rate limit: %w", err)
	}
	if !claimed {
		slog.WarnContext(ctx, "Превышен лимит запросов кода",
			slog.String("email", platformlogger.MaskEmail(email)),
		)
		return apierror.ErrCodeRateLimited
	}
	keepRateLimit := false
	defer func() {
		if !keepRateLimit {
			c.redis.Del(ctx, rateKey)
		}
	}()

	codeKey := CodeKey(email)
	attemptsKey := AttemptsKey(email)

	if err := c.redis.Del(ctx, codeKey, attemptsKey).Err(); err != nil {
		return fmt.Errorf("clear previous verification state: %w", err)
	}

	if err := c.redis.Set(ctx, codeKey, code, VerifyCodeTTL).Err(); err != nil {
		return fmt.Errorf("save verification code: %w", err)
	}

	keepRateLimit = true

	return nil
}

func (c *CodeStore) Verify(ctx context.Context, email, code string) error {
	attemptsKey := AttemptsKey(email)
	cooldownKey := CooldownKey(email)
	codeKey := CodeKey(email)
	lockKey := "lock:verify:" + codeKey

	if cnt, err := c.redis.Exists(ctx, cooldownKey).Result(); err != nil {
		return fmt.Errorf("check verification cooldown: %w", err)
	} else if cnt == 1 {
		slog.WarnContext(ctx, "Попытка проверки кода во время блокировки",
			slog.String("email", platformlogger.MaskEmail(email)),
		)
		return apierror.ErrTooManyAttempts
	}

	ok, err := c.redis.SetNX(ctx, lockKey, "1", 5*time.Second).Result()
	if err != nil {
		return fmt.Errorf("acquire verification lock: %w", err)
	}
	if !ok {
		slog.WarnContext(ctx, "Параллельная попытка проверки кода",
			slog.String("email", platformlogger.MaskEmail(email)),
		)
		return apierror.ErrInvalidVerificationCode
	}

	defer c.redis.Del(ctx, lockKey)

	storedCode, err := c.redis.Get(ctx, codeKey).Result()
	if err != nil {
		if err == redis.Nil {
			slog.WarnContext(ctx, "Попытка проверки истёкшего кода",
				slog.String("email", platformlogger.MaskEmail(email)),
			)
			return apierror.ErrCodeExpired
		}
		return fmt.Errorf("get verification code: %w", err)
	}

	if storedCode != code {
		attempts, err := c.redis.Incr(ctx, attemptsKey).Result()
		if err != nil {
			return fmt.Errorf("increment verification attempts: %w", err)
		}

		if attempts == 1 {
			if err := c.redis.Expire(ctx, attemptsKey, VerifyCodeTTL).Err(); err != nil {
				return fmt.Errorf("set verification attempts ttl: %w", err)
			}
		}

		if attempts >= MaxVerifyAttempts {
			slog.WarnContext(ctx, "Превышен лимит попыток проверки кода",
				slog.String("email", platformlogger.MaskEmail(email)),
				slog.Int64("attempts", attempts),
			)
			if err := c.redis.Set(ctx, cooldownKey, "1", CooldownAfterFailed).Err(); err != nil {
				return fmt.Errorf("set verification cooldown: %w", err)
			}
			c.redis.Del(ctx, codeKey, attemptsKey)

			return apierror.ErrTooManyAttempts
		}

		slog.WarnContext(ctx, "Неверный код подтверждения",
			slog.String("email", platformlogger.MaskEmail(email)),
			slog.Int64("attempt", attempts),
			slog.Int("max_attempts", MaxVerifyAttempts),
		)
		return apierror.ErrInvalidVerificationCode
	}

	if err := c.redis.Del(ctx, codeKey, attemptsKey).Err(); err != nil {
		return fmt.Errorf("invalidate verification code: %w", err)
	}

	slog.DebugContext(ctx, "Код подтверждения и счётчик попыток очищены",
		slog.String("email", platformlogger.MaskEmail(email)),
	)

	return nil
}
