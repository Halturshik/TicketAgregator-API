package service

import (
	"context"

	"github.com/Halturshik/TicketAgregator-API/internal/auth"
	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/common/cleaning"
	"github.com/Halturshik/TicketAgregator-API/internal/common/validator"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
)

func (s *Service) Refresh(ctx context.Context, refreshToken string) (*auth.LoginOutput, error) {
	refreshToken = cleaning.Token(refreshToken)

	fields := map[string]string{}

	if !validator.NotEmpty(refreshToken) {
		fields[apierror.FieldRefreshToken] = apierror.ErrInvalidRefreshToken
	}
	if len(fields) > 0 {
		return nil, apierror.Validation(fields)
	}

	userID, tokenVersion, err := s.jwt.ParseRefreshToken(refreshToken)
	if err != nil {
		logger.Warn("Невалидный refresh-токен при попытке обновления: %v", err)
		return nil, apierror.ErrInvalidToken
	}

	user, err := s.store.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if user.TokenVersion != tokenVersion {
		logger.Warn("Несоответствие версии токена: в JWT %d, в DB %d", tokenVersion, user.TokenVersion)
		return nil, apierror.ErrInvalidToken
	}

	accessToken, err := s.jwt.GenerateAccessToken(userID, user.TokenVersion)
	if err != nil {
		logger.Error("Ошибка при генерации access-токен для userID %v: %v", userID, err)
		return nil, err
	}

	newRefreshToken, err := s.jwt.GenerateRefreshToken(userID, user.TokenVersion)
	if err != nil {
		logger.Error("Ошибка при генерации нового refresh-токена для userID %v: %v", userID, err)
		return nil, err
	}

	if err := s.refreshStore.Rotate(ctx, int64(userID), refreshToken, newRefreshToken); err != nil {
		if err == apierror.ErrInvalidToken {
			logger.Warn("Попытка повторно использовать refresh-токен: userID=%d", userID)
			return nil, apierror.ErrInvalidToken
		}
		return nil, err
	}

	logger.Info("Refresh-токен успешно обновлён для userID: %v", userID)

	return &auth.LoginOutput{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}
