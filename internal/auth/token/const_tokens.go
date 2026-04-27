package token

import "time"

const (
	AccessTokenTTL  = 15 * time.Minute
	RefreshTokenTTL = 14 * 24 * time.Hour
)
