package service

import (
	"context"
	"errors"
	"log/slog"

	"github.com/Halturshik/TicketAgregator-API/internal/auth"
	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/common/cleaning"
	"github.com/Halturshik/TicketAgregator-API/internal/common/validator"
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
		slog.WarnContext(ctx, "Отклонён невалидный refresh-токен", slog.Any("error", err))
		return nil, apierror.ErrInvalidToken
	}

	user, err := s.store.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if user.TokenVersion != tokenVersion {
		slog.WarnContext(ctx, "Несоответствие версии refresh-токена",
			slog.Int("user_id", userID),
			slog.Int("token_version", tokenVersion),
			slog.Int("current_version", user.TokenVersion),
		)
		return nil, apierror.ErrInvalidToken
	}

	accessToken, err := s.jwt.GenerateAccessToken(userID, user.TokenVersion)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := s.jwt.GenerateRefreshToken(userID, user.TokenVersion)
	if err != nil {
		return nil, err
	}

	if err := s.refreshStore.Rotate(ctx, int64(userID), refreshToken, newRefreshToken); err != nil {
		if errors.Is(err, apierror.ErrInvalidToken) {
			slog.WarnContext(ctx, "Попытка повторно использовать refresh-токен", slog.Int("user_id", userID))
			return nil, apierror.ErrInvalidToken
		}
		return nil, err
	}

	slog.InfoContext(ctx, "Refresh-токен успешно обновлён", slog.Int("user_id", userID))

	return &auth.LoginOutput{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}
