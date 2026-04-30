package service

import (
	"context"

	"github.com/Halturshik/TicketAgregator-API/internal/auth"
	"github.com/Halturshik/TicketAgregator-API/internal/auth/repository"
	"github.com/Halturshik/TicketAgregator-API/internal/auth/token"
	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/common/cleaning"
	"github.com/Halturshik/TicketAgregator-API/internal/common/validator"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
)

func (s *Service) StartRegistration(ctx context.Context, in auth.RegisterInput) error {
	in.FirstName = cleaning.Name(in.FirstName)
	in.MiddleName = cleaning.Name(in.MiddleName)
	in.LastName = cleaning.Name(in.LastName)
	in.Email = cleaning.Email(in.Email)
	in.Password = cleaning.Password(in.Password)

	fields := map[string]string{}

	if !validator.NotEmpty(in.FirstName) || !validator.ValidName(in.FirstName) {
		fields[apierror.FieldFirstName] = apierror.ErrInvalidFirstName
	}

	if validator.NotEmpty(in.MiddleName) && !validator.ValidName(in.MiddleName) {
		fields[apierror.FieldMiddleName] = apierror.ErrInvalidMiddleName
	}

	if !validator.NotEmpty(in.LastName) || !validator.ValidName(in.LastName) {
		fields[apierror.FieldLastName] = apierror.ErrInvalidLastName
	}

	if !validator.ValidEmail(in.Email) {
		fields[apierror.FieldEmail] = apierror.ErrInvalidEmail
	}

	if _, err := validator.ValidBirthDate(in.BirthDate); err != nil {
		fields[apierror.FieldBirthDate] = apierror.ErrInvalidBirthDate
	}

	if !validator.ValidPassword(in.Password) {
		fields[apierror.FieldPassword] = apierror.ErrInvalidPassword
	}

	if len(fields) > 0 {
		return apierror.Validation(fields)
	}

	exists, err := s.store.IsEmailExists(ctx, in.Email)
	if err != nil {
		logger.Error("Ошибка при проверке существования email %s: %v", in.Email, err)
		return err
	}
	if exists {
		logger.Warn("Попытка регистрации на уже существующий email: %s", in.Email)
		return apierror.ErrEmailIsUsed
	}

	err = s.registrationStore.Save(ctx, in.Email, in)
	if err != nil {
		return err
	}

	if err := s.codeSender.Send(ctx, in.Email); err != nil {
		return err
	}

	return nil
}

func (s *Service) ConfirmRegistration(ctx context.Context, in auth.ConfirmRegisterInput) (*auth.LoginOutput, error) {
	in.Email = cleaning.Email(in.Email)

	if err := s.codeStore.Verify(ctx, in.Email, in.Code); err != nil {
		return nil, err
	}

	stored, err := s.registrationStore.Get(ctx, in.Email)
	if err != nil {
		return nil, err
	}

	birthDate, _ := validator.ValidBirthDate(stored.BirthDate)

	hashPassword, err := token.HashPassword(stored.Password)
	if err != nil {
		logger.Error("Ошибка при хэшировании пароля для %s: %v", in.Email, err)
		return nil, err
	}

	dbParams := repository.CreateUserParams{
		FirstName:    stored.FirstName,
		MiddleName:   stored.MiddleName,
		LastName:     stored.LastName,
		BirthDate:    birthDate,
		Email:        stored.Email,
		PasswordHash: hashPassword,
		IsRussian:    stored.IsRussian,
	}

	userID, err := s.store.CreateUser(ctx, dbParams)
	if err != nil {
		if err == repository.ErrDuplicateEmail {
			return nil, apierror.ErrEmailIsUsed
		}
		logger.Error("Ошибка при создании пользователя %s: %v", in.Email, err)
		return nil, err
	}

	accessToken, err := s.jwt.GenerateAccessToken(int(userID), 1)
	if err != nil {
		logger.Error("Ошибка при генерации access-токен для userID %v: %v", userID, err)
		return nil, err
	}

	refreshToken, err := s.jwt.GenerateRefreshToken(int(userID), 1)
	if err != nil {
		logger.Error("Ошибка при генерации refresh-токена для userID %v: %v", userID, err)
		return nil, err
	}

	hashToken := token.HashToken(refreshToken)

	if err := s.refreshStore.Save(ctx, userID, refreshToken); err != nil {
		return nil, err
	}

	if err := s.refreshStore.AddToUserSet(ctx, userID, hashToken); err != nil {
		return nil, err
	}

	if err := s.codeStore.Clear(ctx, in.Email); err != nil {
		return nil, err
	}

	if err := s.registrationStore.Delete(ctx, in.Email); err != nil {
		return nil, err
	}

	logger.Info("Успешная регистрация нового пользователя: userID/email: %v/%s", userID, in.Email)

	return &auth.LoginOutput{
		AccessToken: accessToken,
	}, nil
}
