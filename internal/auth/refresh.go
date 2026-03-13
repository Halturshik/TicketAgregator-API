package auth

import (
	"context"

	"github.com/Halturshik/TicketAgregator-API/internal/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/auth/types"
	"github.com/Halturshik/TicketAgregator-API/internal/authutils"
	"github.com/Halturshik/TicketAgregator-API/internal/logger"
)

func (s *Service) Refresh(ctx context.Context, refreshToken string) (*types.LoginOutput, error) {

	userIDFromJWT, err := authutils.ParseToken(refreshToken)
	if err != nil {
		return nil, apierror.ErrInvalidToken
	}

	userID, err := s.refreshStore.Get(ctx, refreshToken)
	if err != nil {
		logger.Error("Ошибка при получения refresh-токена из redis для userID %v: %v", userID, err)
		return nil, apierror.ErrInvalidToken
	}

	if int64(userIDFromJWT) != userID {
		logger.Warn("Несоответствие userID в refresh токене")
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
