package bookingaccess

import "time"

const (
	VerificationCodeTTL     = 5 * time.Minute
	CodeRequestInterval     = 2 * time.Minute
	VerificationCooldown    = 5 * time.Minute
	AccessTokenTTL          = 15 * time.Minute
	VerificationLockTTL     = 5 * time.Second
	MaxVerificationAttempts = 5
)
