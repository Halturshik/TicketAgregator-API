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
		return err
	}
	if exists {
		slog.WarnContext(ctx, "Попытка регистрации на существующий email",
			slog.String("email", platformlogger.MaskEmail(in.Email)),
		)
		return apierror.ErrEmailIsUsed
	}

	passwordHash, err := password.HashPassword(in.Password)
	if err != nil {
		return err
	}

	pending := auth.PendingRegistration{
		FirstName:    in.FirstName,
		MiddleName:   in.MiddleName,
		LastName:     in.LastName,
		BirthDate:    in.BirthDate,
		Email:        in.Email,
		PasswordHash: passwordHash,
		IsRussian:    in.IsRussian,
	}

	if err := s.registrationStore.Save(ctx, in.Email, pending); err != nil {
		return err
	}

	code, err := s.codeGenerator.GenerateVerificationCode()
	if err != nil {
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

func (s *Service) ConfirmRegistration(ctx context.Context, in auth.ConfirmRegisterInput) (*auth.LoginOutput, error) {
	in.Email = cleaning.Email(in.Email)

	if err := s.codeStore.Verify(ctx, in.Email, in.Code); err != nil {
		return nil, err
	}

	stored, err := s.registrationStore.Get(ctx, in.Email)
	if err != nil {
		return nil, err
	}
	if stored.PasswordHash == "" {
		return nil, auth.ErrInvalidRegistrationData
	}

	birthDate, _ := validator.ValidBirthDate(stored.BirthDate)

	dbParams := auth.CreateUserParams{
		FirstName:    stored.FirstName,
		MiddleName:   stored.MiddleName,
		LastName:     stored.LastName,
		BirthDate:    birthDate,
		Email:        stored.Email,
		PasswordHash: stored.PasswordHash,
		IsRussian:    stored.IsRussian,
	}

	userID, err := s.store.CreateUser(ctx, dbParams)
	if err != nil {
		if errors.Is(err, auth.ErrDuplicateEmail) {
			return nil, apierror.ErrEmailIsUsed
		}
		return nil, err
	}

	accessToken, err := s.jwt.GenerateAccessToken(int(userID), 1)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwt.GenerateRefreshToken(int(userID), 1)
	if err != nil {
		return nil, err
	}

	hashToken := token.HashToken(refreshToken)

	if err := s.refreshStore.Save(ctx, userID, refreshToken); err != nil {
		return nil, err
	}

	if err := s.refreshStore.AddToUserSet(ctx, userID, hashToken); err != nil {
		return nil, err
	}

	if err := s.registrationStore.Delete(ctx, in.Email); err != nil {
		return nil, err
	}

	slog.InfoContext(ctx, "Пользователь успешно зарегистрирован",
		slog.Int64("user_id", userID),
		slog.String("email", platformlogger.MaskEmail(in.Email)),
	)

	return &auth.LoginOutput{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
