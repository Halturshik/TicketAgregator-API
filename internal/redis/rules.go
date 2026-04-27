package redis

import "time"

const (
	MaxVerifyAttempts = 5

	VerifyCodeTTL       = 5 * time.Minute
	CodeRateLimitWindow = 2 * time.Minute
	CooldownAfterFailed = 5 * time.Minute

	RegistrationTTL  = 10 * time.Minute
	LoginTTL         = 10 * time.Minute
	ResetPasswordTTL = 10 * time.Minute
)
