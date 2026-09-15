package service

import (
	"context"
	"errors"
	"log/slog"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/common/cleaning"
	"github.com/Halturshik/TicketAgregator-API/internal/common/validator"
	"github.com/Halturshik/TicketAgregator-API/internal/users"
)

func (s *Service) UpdateProfile(ctx context.Context, userID int, in users.UpdateProfileInput) (*users.Profile, error) {
	in.FirstName = cleaning.Name(in.FirstName)
	in.MiddleName = cleaning.Name(in.MiddleName)
	in.LastName = cleaning.Name(in.LastName)

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
	birthDate, err := validator.ValidBirthDate(in.BirthDate)
	if err != nil {
		fields[apierror.FieldBirthDate] = apierror.ErrInvalidBirthDate
	}
	if len(fields) > 0 {
		return nil, apierror.Validation(fields)
	}

	profile, err := s.repo.UpdateProfile(ctx, userID, in, birthDate)
	if errors.Is(err, users.ErrNotFound) {
		return nil, apierror.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	slog.InfoContext(ctx, "Профиль пользователя обновлён", slog.Int("user_id", userID))
	return profile, nil
}
