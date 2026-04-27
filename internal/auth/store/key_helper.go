package store

import "fmt"

func RegistrationKey(email string) string {
	return KeyAuthRegistration + email
}

func LoginKey(email string) string {
	return KeyAuthLogin + email
}

func CodeKey(email string) string {
	return KeyAuthCode + email
}

func AttemptsKey(email string) string {
	return KeyAuthAttempts + email
}

func RateLimitKey(email string) string {
	return KeyAuthRateLimit + email
}

func CooldownKey(email string) string {
	return KeyAuthCooldown + email
}

func RefreshKey(token string) string {
	return KeyAuthRefresh + token
}

func ResetVerifiedKey(email string) string {
	return KeyAuthResetVerified + email
}

func RefreshUserSetKey(userID int64) string {
	return fmt.Sprintf("%suser:%d", KeyAuthRefresh, userID)
}
