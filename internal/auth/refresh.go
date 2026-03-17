package auth

import (
	"context"

	"github.com/Halturshik/TicketAgregator-API/internal/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/auth/types"
	"github.com/Halturshik/TicketAgregator-API/internal/authutils"
	"github.com/Halturshik/TicketAgregator-API/internal/cleaning"
	"github.com/Halturshik/TicketAgregator-API/internal/logger"
	"github.com/Halturshik/TicketAgregator-API/internal/validator"
)

func (s *Service) Refresh(ctx context.Context, refreshToken string) (*types.LoginOutput, error) {
	refreshToken = cleaning.Token(refreshToken)

	fields := map[string]string{}

	if !validator.NotEmpty(refreshToken) {
		fields[apierror.FieldRefreshToken] = apierror.ErrInvalidRefreshToken
	}
	if len(fields) > 0 {
		return nil, apierror.Validation(fields)
	}

	userIDFromJWT, err := authutils.ParseToken(refreshToken)
	if err != nil {
		logger.Warn("Невалидный refresh-токен при попытке обновления: %v", err)
		return nil, apierror.ErrInvalidToken
	}

	userID, err := s.refreshStore.Get(ctx, refreshToken)
	if err != nil {
		if err == apierror.ErrInvalidToken {
			logger.Warn("Попытка использовать отсутствующий/истёкший refresh-токен (userID из JWT: %d)", userIDFromJWT)
			return nil, apierror.ErrInvalidToken
		}

		logger.Error("Ошибка при получения refresh-токена из redis (userID из JWT: %d): %v", userIDFromJWT, err)
		return nil, apierror.ErrInternal
	}

	if int64(userIDFromJWT) != userID {
		logger.Warn("Несоответствие userID: в JWT %d, в Redis %d", userIDFromJWT, userID)
		return nil, apierror.ErrInvalidToken
	}

	if err := s.refreshStore.Delete(ctx, refreshToken); err != nil {
		logger.Error("Ошибка при удалении refresh-токена из redis для userID %v: %v", userID, err)
		return nil, err
	}

	accessToken, err := authutils.GenerateToken(int(userID), authutils.AccessTokenTTL)
	if err != nil {
		logger.Error("Ошибка при генерации access-токен для userID %v: %v", userID, err)
		return nil, err
	}
	logger.Info("Сгенерирован access-токена для userID %v", userID)

	newRefreshToken, err := authutils.GenerateToken(int(userID), authutils.RefreshTokenTTL)
	if err != nil {
		logger.Error("Ошибка при генерации нового refresh-токена для userID %v: %v", userID, err)
		return nil, err
	}
	logger.Info("Сгенерирован новый refresh-токен для userID %v", userID)

	if err := s.refreshStore.Save(ctx, userID, newRefreshToken); err != nil {
		return nil, err
	}

	return &types.LoginOutput{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}
