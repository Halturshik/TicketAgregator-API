package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/common/httpx"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
)

type contextKey struct{}

var userIDKey = contextKey{}

type AuthMiddleware struct {
	store UserStore
	jwt   TokenManager
}

func NewAuthMiddleware(store UserStore, jwt TokenManager) *AuthMiddleware {
	return &AuthMiddleware{
		store: store,
		jwt:   jwt,
	}
}

func (m *AuthMiddleware) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			logger.Warn("Запрос без заголовка Authorization: %s %s", r.Method, r.URL.Path)
			_ = httpx.WriteJSON(w, apierror.ErrUnauthorized.Status, apierror.ErrUnauthorized)
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			logger.Warn("Некорректный формат токена в заголовке Authorization: %s %s", r.Method, r.URL.Path)
			_ = httpx.WriteJSON(w, apierror.ErrInvalidTokenFormat.Status, apierror.ErrInvalidTokenFormat)
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		userID, tokenVersion, err := m.jwt.ParseToken(tokenStr)
		if err != nil {
			logger.Warn("Ошибка при разборе access-токена: %s %s", r.Method, r.URL.Path)
			_ = httpx.WriteJSON(w, apierror.ErrInvalidToken.Status, apierror.ErrInvalidToken)
			return
		}

		user, err := m.store.GetUserByID(r.Context(), userID)
		if err != nil {
			logger.Warn("Пользователь не найден по токену: userID: %d", userID)
			_ = httpx.WriteJSON(w, apierror.ErrUnauthorized.Status, apierror.ErrUnauthorized)
			return
		}

		if user.TokenVersion != tokenVersion {
			logger.Warn("Несовпадение версии токена для userID: %d", userID)
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
