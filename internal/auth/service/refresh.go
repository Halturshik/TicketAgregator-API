package service

import (
	"context"

	"github.com/Halturshik/TicketAgregator-API/internal/auth"
	"github.com/Halturshik/TicketAgregator-API/internal/auth/token"
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

	userIDFromJWT, tokenVersion, err := s.jwt.ParseToken(refreshToken)
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

		return nil, apierror.ErrInternal
	}

	if int64(userIDFromJWT) != userID {
		logger.Warn("Несоответствие userID: в JWT %d, в Redis %d", userIDFromJWT, userID)
		return nil, apierror.ErrInvalidToken
	}

	user, err := s.store.GetUserByID(ctx, userIDFromJWT)
	if err != nil {
		return nil, err
	}

	if user.TokenVersion != tokenVersion {
		logger.Warn("Несоответствие версии токена: в JWT %d, в DB %d", tokenVersion, user.TokenVersion)
		return nil, apierror.ErrInvalidToken
	}

	if err := s.refreshStore.Delete(ctx, refreshToken); err != nil {
		return nil, err
	}

	oldHash := token.HashToken(refreshToken)

	if err := s.refreshStore.RemoveFromUserSet(ctx, userID, oldHash); err != nil {
	}

	accessToken, err := s.jwt.GenerateAccessToken(int(userID), user.TokenVersion)
	if err != nil {
		logger.Error("Ошибка при генерации access-токен для userID %v: %v", userID, err)
		return nil, err
	}

	newRefreshToken, err := s.jwt.GenerateRefreshToken(int(userID), user.TokenVersion)
	if err != nil {
		logger.Error("Ошибка при генерации нового refresh-токена для userID %v: %v", userID, err)
		return nil, err
	}

	hash := token.HashToken(newRefreshToken)

	if err := s.refreshStore.Save(ctx, userID, newRefreshToken); err != nil {
		return nil, err
	}

	if err := s.refreshStore.AddToUserSet(ctx, userID, hash); err != nil {
		return nil, err
	}

	logger.Info("Refresh-токен успешно обновлён для userID: %v", userIDFromJWT)

	return &auth.LoginOutput{
		AccessToken: accessToken,
	}, nil
}
