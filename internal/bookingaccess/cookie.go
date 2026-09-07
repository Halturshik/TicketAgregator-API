package bookingaccess

import (
	"net/http"
	"time"
)

const (
	AccessCookieName = "booking_access_token"
	accessCookiePath = "/api/bookings"
)

func SetAccessToken(w http.ResponseWriter, token string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name: AccessCookieName, Value: token, Path: accessCookiePath,
		HttpOnly: true, Secure: false, SameSite: http.SameSiteLaxMode,
		MaxAge: int(AccessTokenTTL.Seconds()), Expires: expiresAt,
	})
}

func AccessToken(r *http.Request) string {
	cookie, err := r.Cookie(AccessCookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}
