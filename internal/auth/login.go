package auth

import (
	"context"
	"errors"

	"github.com/Halturshik/TicketAgregator-API/database/errs"
	"github.com/Halturshik/TicketAgregator-API/internal/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/auth/types"
	"github.com/Halturshik/TicketAgregator-API/internal/authutils"
	"github.com/Halturshik/TicketAgregator-API/internal/cleaning"
	"github.com/Halturshik/TicketAgregator-API/internal/logger"
	"github.com/Halturshik/TicketAgregator-API/internal/validator"
)

func (s *Service) LoginStart(ctx context.Context, in types.LoginStartInput) error {
	in.Email = cleaning.Email(in.Email)
	in.Password = cleaning.Password(in.Password)

	fields := map[string]string{}

	if !validator.ValidEmail(in.Email) {
		fields[apierror.FieldEmail] = apierror.ErrInvalidEmail
	}

	if !validator.ValidPassword(in.Password) {
		fields[apierror.FieldPassword] = apierror.ErrInvalidPassword
	}

	if len(fields) > 0 {
		return apierror.Validation(fields)
	}

	user, err := s.store.GetUserByEmail(ctx, in.Email)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			return apierror.ErrInvalidCredentials
		}
		logger.Error("Ошибка при получении пользователя по email %s: %v", in.Email, err)
		return err
	}

	if err := authutils.CheckPassword(user.PasswordHash, in.Password); err != nil {
		return apierror.ErrInvalidCredentials
	}

	if err := s.loginStore.Save(ctx, in.Email, user.ID); err != nil {
		return err
	}

	if err := s.codeSender.Send(ctx, in.Email); err != nil {
		return err
	}

	return nil
}

func (s *Service) LoginConfirm(ctx context.Context, in types.LoginConfirmInput) (*types.LoginOutput, error) {
	in.Email = cleaning.Email(in.Email)

	if err := s.codeStore.Verify(ctx, in.Email, in.Code); err != nil {
		return nil, err
	}

	userID, err := s.loginStore.Get(ctx, in.Email)
	if err != nil {
		return nil, err
	}

	accessToken, err := authutils.GenerateToken(int(userID), authutils.AccessTokenTTL)
	if err != nil {
		logger.Error("Ошибка при генерации access-токен для userID %v: %v", userID, err)
		return nil, err
	}
	logger.Info("Сгенерирован access-токена для userID %v", userID)

	refreshToken, err := authutils.GenerateToken(int(userID), authutils.RefreshTokenTTL)
	if err != nil {
		logger.Error("Ошибка при генерации refresh-токена для userID %v: %v", userID, err)
		return nil, err
	}
	logger.Info("Сгенерирован refresh-токен для userID %v", userID)

	if err := s.refreshStore.Save(ctx, userID, refreshToken); err != nil {
		return nil, err
	}

	if err := s.codeStore.Clear(ctx, in.Email); err != nil {
		return nil, err
	}

	if err := s.loginStore.Delete(ctx, in.Email); err != nil {
		return nil, err
	}

	return &types.LoginOutput{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
