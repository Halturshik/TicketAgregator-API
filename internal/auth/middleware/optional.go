package middleware

import (
	"net/http"

	"github.com/Halturshik/TicketAgregator-API/internal/auth"
)

func (m *Middleware) Optional(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization := r.Header.Get("Authorization")
		if authorization == "" {
			next.ServeHTTP(w, r)
			return
		}

		userID, err := m.authenticate(r.Context(), authorization)
		if err != nil {
			writeError(w, r, err)
			return
		}

		next.ServeHTTP(w, r.WithContext(auth.WithUserID(r.Context(), userID)))
	})
}
