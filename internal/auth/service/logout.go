package service

import (
	"context"

	"github.com/Halturshik/TicketAgregator-API/internal/auth/token"
	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/common/cleaning"
	"github.com/Halturshik/TicketAgregator-API/internal/common/validator"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
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

	userIDFromJWT, _, err := token.ParseToken(refreshToken)
	if err != nil {
		if err := s.refreshStore.Delete(ctx, refreshToken); err != nil {
			logger.Error("Ошибка при удалении битого refresh-токена (%s) из redis при logout: %v", refreshToken[:8], err)
			return err
		}

		return nil
	}

	hash := token.HashToken(refreshToken)

	if err := s.refreshStore.Delete(ctx, refreshToken); err != nil {
		logger.Error("Ошибка при удаления refresh-токена (%s) из redis при logout для userID %d: %v", refreshToken[:8], userIDFromJWT, err)
		return err
	}

	if err := s.refreshStore.RemoveFromUserSet(ctx, int64(userIDFromJWT), hash); err != nil {
	}

	logger.Info("Пользователь успешно разлогинен: userID %d", userIDFromJWT)
	return nil
}
