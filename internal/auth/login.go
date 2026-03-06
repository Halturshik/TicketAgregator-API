package auth

import (
	"context"
	"errors"
	"time"

	"github.com/Halturshik/TicketAgregator-API/database/errs"
	"github.com/Halturshik/TicketAgregator-API/internal/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/cleaning"
	"github.com/Halturshik/TicketAgregator-API/internal/validator"
	"golang.org/x/crypto/bcrypt"
)

type LoginStartInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginConfirmInput struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type LoginOutput struct {
	Token string `json:"token"`
}

func (s *Service) LoginStart(ctx context.Context, in LoginStartInput) error {
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
		return err
	}

	if bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(in.Password),
	) != nil {
		return apierror.ErrInvalidCredentials
	}

	code, err := s.codeService.Generate(ctx, in.Email)
	if err != nil {
		return err
	}

	return s.mailer.SendVerificationEmail(in.Email, code)
}

func (s *Service) LoginConfirm(ctx context.Context, in LoginConfirmInput) (*LoginOutput, error) {
	in.Email = cleaning.Email(in.Email)

	if err := s.codeService.Verify(ctx, in.Email, in.Code); err != nil {
		return nil, err
	}

	user, err := s.store.GetUserByEmail(ctx, in.Email)
	if err != nil {
		return nil, err
	}

	token, err := GenerateToken(user.ID, 30*time.Minute)
	if err != nil {
		return nil, err
	}

	return &LoginOutput{
		Token: token,
	}, nil
}
