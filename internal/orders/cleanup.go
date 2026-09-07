package orders

import "time"

const (
	ExpiredOrderRetention = 24 * time.Hour
	DefaultCleanupLimit   = 500
	MaxCleanupLimit       = 2000
)

func NormalizeCleanupLimit(limit int) int {
	if limit <= 0 {
		return DefaultCleanupLimit
	}
	if limit > MaxCleanupLimit {
		return MaxCleanupLimit
	}
	return limit
}
