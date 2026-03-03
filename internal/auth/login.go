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

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginOutput struct {
	Token string `json:"token"`
}

func (s *Service) Login(ctx context.Context, in LoginInput) (*LoginOutput, error) {
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

	user, err := s.store.GetUserByEmail(ctx, in.Email)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			return nil, apierror.ErrInvalidCredentials
		}
		return nil, err
	}

	if bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(in.Password),
	) != nil {
		return nil, apierror.ErrInvalidCredentials
	}

	token, err := GenerateToken(user.ID, 30*time.Minute)
	if err != nil {
		return nil, err
	}

	return &LoginOutput{
		Token: token,
	}, nil
}
