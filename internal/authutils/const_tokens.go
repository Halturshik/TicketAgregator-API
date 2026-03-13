package authutils

import "time"

const (
	AccessTokenTTL  = 30 * time.Minute
	RefreshTokenTTL = 14 * 24 * time.Hour
)
