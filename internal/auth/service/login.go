package service

import (
	"context"
	"errors"

	"github.com/Halturshik/TicketAgregator-API/internal/auth"
	"github.com/Halturshik/TicketAgregator-API/internal/auth/password"
	"github.com/Halturshik/TicketAgregator-API/internal/auth/repository"
	"github.com/Halturshik/TicketAgregator-API/internal/auth/token"
	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/common/cleaning"
	"github.com/Halturshik/TicketAgregator-API/internal/common/validator"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
)

func (s *Service) LoginStart(ctx context.Context, in auth.LoginStartInput) error {
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
		if errors.Is(err, repository.ErrUserNotFound) {
			return apierror.ErrInvalidCredentials
		}
		logger.Error("Ошибка при получении пользователя по email %s: %v", in.Email, err)
		return err
	}

	if err := password.CheckPassword(user.PasswordHash, in.Password); err != nil {
		return apierror.ErrInvalidCredentials
	}

	if err := s.loginStore.Save(ctx, in.Email, int64(user.ID)); err != nil {
		return err
	}

	code, err := s.codeGenerator.GenerateVerificationCode()
	if err != nil {
		logger.Warn("Ошибка при генерации кода для %s: %v", in.Email, err)
		return err
	}

	if err := s.codeStore.RequestCode(ctx, in.Email, code); err != nil {
		return err
	}

	if err := s.mailer.SendVerificationEmail(ctx, in.Email, code); err != nil {
		return err
	}

	return nil
}

func (s *Service) LoginConfirm(ctx context.Context, in auth.LoginConfirmInput) (*auth.LoginOutput, error) {
	in.Email = cleaning.Email(in.Email)

	if err := s.codeStore.Verify(ctx, in.Email, in.Code); err != nil {
		return nil, err
	}

	userID, err := s.loginStore.Get(ctx, in.Email)
	if err != nil {
		return nil, err
	}

	user, err := s.store.GetUserByEmail(ctx, in.Email)
	if err != nil {
		return nil, err
	}

	accessToken, err := s.jwt.GenerateAccessToken(int(userID), user.TokenVersion)
	if err != nil {
		logger.Error("Ошибка при генерации access-токен для userID %v: %v", userID, err)
		return nil, err
	}

	refreshToken, err := s.jwt.GenerateRefreshToken(int(userID), user.TokenVersion)
	if err != nil {
		logger.Error("Ошибка при генерации refresh-токена для userID %v: %v", userID, err)
		return nil, err
	}

	hash := token.HashToken(refreshToken)

	if err := s.refreshStore.Save(ctx, userID, refreshToken); err != nil {
		return nil, err
	}

	if err := s.refreshStore.AddToUserSet(ctx, userID, hash); err != nil {
		return nil, err
	}

	if err := s.loginStore.Delete(ctx, in.Email); err != nil {
		return nil, err
	}

	logger.Info("Успешная авторизация для userID/email: %v/%s", userID, in.Email)

	return &auth.LoginOutput{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
