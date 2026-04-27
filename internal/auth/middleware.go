package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/Halturshik/TicketAgregator-API/internal/auth/token"
	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/common/httpx"
)

type contextKey struct{}

var userIDKey = contextKey{}

type AuthMiddleware struct {
	store UserStore
}

func NewAuthMiddleware(store UserStore) *AuthMiddleware {
	return &AuthMiddleware{store: store}
}

func (m *AuthMiddleware) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			_ = httpx.WriteJSON(w, apierror.ErrUnauthorized.Status, apierror.ErrUnauthorized)
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			_ = httpx.WriteJSON(w, apierror.ErrInvalidTokenFormat.Status, apierror.ErrInvalidTokenFormat)
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		userID, tokenVersion, err := token.ParseToken(tokenStr)
		if err != nil {
			_ = httpx.WriteJSON(w, apierror.ErrInvalidToken.Status, apierror.ErrInvalidToken)
			return
		}

		user, err := m.store.GetUserByID(r.Context(), userID)
		if err != nil {
			_ = httpx.WriteJSON(w, apierror.ErrUnauthorized.Status, apierror.ErrUnauthorized)
			return
		}

		if user.TokenVersion != tokenVersion {
			_ = httpx.WriteJSON(w, apierror.ErrInvalidToken.Status, apierror.ErrInvalidToken)
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func UserIDFromContext(ctx context.Context) (int, bool) {
	val := ctx.Value(userIDKey)
	if val == nil {
		return 0, false
	}
	id, ok := val.(int)
	return id, ok
}
