package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const keyPrefix = "rate_limit:"

var fixedWindowScript = redis.NewScript(`
local current = redis.call('INCR', KEYS[1])
if current == 1 then
  redis.call('PEXPIRE', KEYS[1], ARGV[1])
end
local ttl = redis.call('PTTL', KEYS[1])
return {current, ttl}
`)

type Limiter struct {
	redis *redis.Client
}

func New(redisClient *redis.Client) *Limiter {
	return &Limiter{redis: redisClient}
}

func (l *Limiter) Allow(ctx context.Context, policy Policy, identity string) (Decision, error) {
	if policy.Name == "" || policy.Limit <= 0 || policy.Window <= 0 || identity == "" {
		return Decision{}, fmt.Errorf("invalid rate limit policy")
	}
	result, err := fixedWindowScript.Run(
		ctx,
		l.redis,
		[]string{keyPrefix + policy.Name + ":" + identity},
		policy.Window.Milliseconds(),
	).Int64Slice()
	if err != nil {
		return Decision{}, fmt.Errorf("apply rate limit: %w", err)
	}
	if len(result) != 2 {
		return Decision{}, fmt.Errorf("invalid rate limit response")
	}
	retryAfter := time.Duration(result[1]) * time.Millisecond
	if retryAfter < 0 {
		retryAfter = policy.Window
	}
	return Decision{Allowed: result[0] <= policy.Limit, RetryAfter: retryAfter}, nil
}
