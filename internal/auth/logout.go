package auth

import (
	"context"

	"github.com/Halturshik/TicketAgregator-API/internal/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/cleaning"
	"github.com/Halturshik/TicketAgregator-API/internal/logger"
	"github.com/Halturshik/TicketAgregator-API/internal/validator"
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

	if err := s.refreshStore.Delete(ctx, refreshToken); err != nil {
		logger.Error("Ошибка при удалении refresh-токена (%s) из redis при logout: %v", refreshToken[:8], err)
		return err
	}

	logger.Info("Пользователь успешно разлогинен")
	return nil
}
