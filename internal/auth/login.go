package auth

import (
	"context"
	"errors"
	"time"

	"github.com/Halturshik/TicketAgregator-API/database/errs"
	"github.com/Halturshik/TicketAgregator-API/internal/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/auth/types"
	"github.com/Halturshik/TicketAgregator-API/internal/authutils"
	"github.com/Halturshik/TicketAgregator-API/internal/cleaning"
	"github.com/Halturshik/TicketAgregator-API/internal/validator"
	"golang.org/x/crypto/bcrypt"
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
		return err
	}

	if bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(in.Password),
	) != nil {
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

	token, err := authutils.GenerateToken(int(userID), 30*time.Minute)
	if err != nil {
		return nil, err
	}

	if err := s.codeStore.Clear(ctx, in.Email); err != nil {
		return nil, err
	}

	if err := s.loginStore.Delete(ctx, in.Email); err != nil {
		return nil, err
	}

	return &types.LoginOutput{
		Token: token,
	}, nil
}
