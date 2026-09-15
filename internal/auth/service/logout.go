package service

import (
	"context"
	"log/slog"

	"github.com/Halturshik/TicketAgregator-API/internal/auth/token"
	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/common/cleaning"
	"github.com/Halturshik/TicketAgregator-API/internal/common/validator"
)

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	refreshToken = cleaning.Token(refreshToken)

	fields := map[string]string{}

	if !validator.NotEmpty(refreshToken) {
		fields[apierror.FieldRefreshToken] = apierror.ErrInvalidRefreshToken
	}

	if len(fields) > 0 {
		return apierror.Validation(fields)
	}

	userIDFromJWT, _, err := s.jwt.ParseRefreshToken(refreshToken)
	if err != nil {
		if err := s.refreshStore.Delete(ctx, refreshToken); err != nil {
			return err
		}

		return nil
	}

	hash := token.HashToken(refreshToken)

	if err := s.refreshStore.Delete(ctx, refreshToken); err != nil {
		return err
	}

	if err := s.refreshStore.RemoveFromUserSet(ctx, int64(userIDFromJWT), hash); err != nil {
	}

	slog.InfoContext(ctx, "Пользователь успешно вышел из системы", slog.Int("user_id", userIDFromJWT))
	return nil
}
