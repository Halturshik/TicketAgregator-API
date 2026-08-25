package middleware

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Halturshik/TicketAgregator-API/internal/auth"
	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
)

func (m *Middleware) authenticate(ctx context.Context, authorization string) (int, error) {
	token, err := bearerToken(authorization)
	if err != nil {
		return 0, err
	}

	userID, tokenVersion, err := m.tokens.ParseAccessToken(token)
	if err != nil {
		return 0, apierror.ErrInvalidToken
	}

	user, err := m.users.GetUserByID(ctx, userID)
	if errors.Is(err, auth.ErrUserNotFound) {
		return 0, apierror.ErrUnauthorized
	}
	if err != nil {
		return 0, fmt.Errorf("get authenticated user: %w", err)
	}
	if user == nil {
		return 0, errors.New("authenticated user reader returned nil user")
	}
	if user.TokenVersion != tokenVersion {
		return 0, apierror.ErrInvalidToken
	}

	return userID, nil
}

func bearerToken(authorization string) (string, error) {
	parts := strings.Fields(authorization)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", apierror.ErrInvalidTokenFormat
	}
	return parts[1], nil
}
