package auth

import (
	"net/http"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/auth/token"
)

const (
	RefreshCookieName = "refresh_token"
	RefreshCookiePath = "/api/auth"
)

func SetRefreshToken(w http.ResponseWriter, refreshToken string) {
	http.SetCookie(w, &http.Cookie{
		Name:     RefreshCookieName,
		Value:    refreshToken,
		Path:     RefreshCookiePath,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(token.RefreshTokenTTL.Seconds()),
		Expires:  time.Now().Add(token.RefreshTokenTTL),
	})
}

func ClearRefreshToken(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     RefreshCookieName,
		Value:    "",
		Path:     RefreshCookiePath,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})
}

func GetRefreshToken(r *http.Request) (string, error) {
	cookie, err := r.Cookie(RefreshCookieName)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}
