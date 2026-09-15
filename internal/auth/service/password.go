package service

import (
	"context"
	"errors"
	"log/slog"

	"github.com/Halturshik/TicketAgregator-API/internal/auth"
	"github.com/Halturshik/TicketAgregator-API/internal/auth/password"
	"github.com/Halturshik/TicketAgregator-API/internal/auth/token"
	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/common/cleaning"
	"github.com/Halturshik/TicketAgregator-API/internal/common/validator"
	platformlogger "github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
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
		return err
	}

	if !exists {
		slog.WarnContext(ctx, "Запрошено восстановление пароля для незарегистрированного email",
			slog.String("email", platformlogger.MaskEmail(email)),
		)
		return nil
	}

	code, err := s.codeGenerator.GenerateVerificationCode()
	if err != nil {
		return err
	}

	if err := s.codeStore.RequestCode(ctx, email, code); err != nil {
		return err
	}

	if err := s.mailer.SendVerificationEmail(ctx, email, code); err != nil {
		return err
	}

	return nil
}

func (s *Service) VerifyResetCode(ctx context.Context, in auth.PasswordVerifyInput) error {
	in.Email = cleaning.Email(in.Email)

	if err := s.codeStore.Verify(ctx, in.Email, in.Code); err != nil {
		return err
	}

	if err := s.resetPasswordStore.SaveVerified(ctx, in.Email); err != nil {
		return err
	}

	return nil
}

func (s *Service) ResetPassword(ctx context.Context, in auth.ChangePasswordConfirmInput) (*auth.LoginOutput, error) {
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
		slog.WarnContext(ctx, "Попытка смены пароля без подтверждённого кода",
			slog.String("email", platformlogger.MaskEmail(in.Email)),
		)
		return nil, apierror.ErrUnauthorized
	}

	user, err := s.store.GetUserByEmail(ctx, in.Email)
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			return nil, apierror.ErrInvalidCredentials
		}
		return nil, err
	}

	if err := password.CheckPassword(user.PasswordHash, in.Password); err == nil {
		return nil, apierror.ErrSamePassword
	}

	hashPassword, err := password.HashPassword(in.Password)
	if err != nil {
		return nil, err
	}

	consumed, err := s.resetPasswordStore.ConsumeVerified(ctx, in.Email)
	if err != nil {
		return nil, err
	}
	if !consumed {
		slog.WarnContext(ctx, "Повторная попытка использовать подтверждение для смены пароля",
			slog.String("email", platformlogger.MaskEmail(in.Email)),
		)
		return nil, apierror.ErrUnauthorized
	}

	newVersion, err := s.store.UpdatePassword(ctx, user.ID, hashPassword)
	if err != nil {
		return nil, err
	}

	slog.InfoContext(ctx, "Пароль пользователя изменён",
		slog.Int("user_id", user.ID),
		slog.String("email", platformlogger.MaskEmail(in.Email)),
	)

	if err := s.refreshStore.DeleteAllForUser(ctx, int64(user.ID)); err != nil {
		slog.ErrorContext(ctx, "Не удалось удалить прежние refresh-токены после смены пароля",
			slog.Int("user_id", user.ID),
			slog.Any("error", err),
		)
	}

	accessToken, err := s.jwt.GenerateAccessToken(user.ID, newVersion)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwt.GenerateRefreshToken(user.ID, newVersion)
	if err != nil {
		return nil, err
	}

	if err := s.refreshStore.Save(ctx, int64(user.ID), refreshToken); err != nil {
		return nil, err
	}

	hashToken := token.HashToken(refreshToken)

	if err := s.refreshStore.AddToUserSet(ctx, int64(user.ID), hashToken); err != nil {
		slog.ErrorContext(ctx, "Не удалось добавить новый refresh-токен в список пользователя после смены пароля",
			slog.Int("user_id", user.ID),
			slog.Any("error", err),
		)
	}

	return &auth.LoginOutput{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
