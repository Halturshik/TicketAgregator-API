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

func (s *Service) ForgotPassword(ctx context.Context, email string) error {
	email = cleaning.Email(email)

	fields := map[string]string{}

	if !validator.ValidEmail(email) {
		fields[apierror.FieldEmail] = apierror.ErrInvalidEmail
	}

	if len(fields) > 0 {
		return apierror.Validation(fields)
	}

	exists, err := s.store.IsEmailExists(ctx, email)
	if err != nil {
		logger.Error("Ошибка при проверке существования email %s: %v", email, err)
		return err
	}

	if !exists {
		logger.Warn("Попытка смены пароля у незарегистрированного email: %s", email)
		return nil
	}

	return s.codeSender.Send(ctx, email)
}

func (s *Service) VerifyResetCode(ctx context.Context, in types.PasswordVerifyInput) error {
	in.Email = cleaning.Email(in.Email)

	if err := s.codeStore.Verify(ctx, in.Email, in.Code); err != nil {
		return err
	}

	if err := s.resetPasswordStore.SaveVerified(ctx, in.Email); err != nil {
		return err
	}

	if err := s.codeStore.Clear(ctx, in.Email); err != nil {
		logger.Warn("Ошибка при инвалидации кода после смены пароля для %s: %v", in.Email, err)
	}

	return nil
}

func (s *Service) ResetPassword(ctx context.Context, in types.ChangePasswordConfirmInput) (*types.LoginOutput, error) {
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
		return nil, apierror.Validation(fields)
	}

	ok, err := s.resetPasswordStore.IsVerified(ctx, in.Email)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, apierror.ErrUnauthorized
	}

	user, err := s.store.GetUserByEmail(ctx, in.Email)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			return nil, apierror.ErrInvalidCredentials
		}
		logger.Error("Ошибка при получении пользователя по email %s: %v", in.Email, err)
		return nil, err
	}

	hashPassword, err := authutils.HashPassword(in.Password)
	if err != nil {
		logger.Error("Ошибка при хэшировании пароля для %s: %v", in.Email, err)
		return nil, err
	}

	newVersion, err := s.store.UpdatePassword(ctx, user.ID, hashPassword)
	if err != nil {
		logger.Error("Ошибка при обновлении пароля в БД для %s: %v", in.Email, err)
		return nil, err
	}

	if err := s.refreshStore.DeleteAllForUser(ctx, int64(user.ID)); err != nil {
	}

	accessToken, err := authutils.GenerateAccessToken(user.ID, newVersion)
	if err != nil {
		return nil, err
	}

	refreshToken, err := authutils.GenerateRefreshToken(user.ID, newVersion)
	if err != nil {
		return nil, err
	}

	if err := s.refreshStore.Save(ctx, int64(user.ID), refreshToken); err != nil {
		return nil, err
	}

	hashToken := authutils.HashToken(refreshToken)

	if err := s.refreshStore.AddToUserSet(ctx, int64(user.ID), hashToken); err != nil {
	}

	if err := s.resetPasswordStore.Delete(ctx, in.Email); err != nil {
	}

	return &types.LoginOutput{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
