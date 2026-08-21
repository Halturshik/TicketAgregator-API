package service

import (
	"context"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/common/cleaning"
	"github.com/Halturshik/TicketAgregator-API/internal/common/validator"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
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
	if err != nil {
		logger.Error("Ошибка при обновлении профиля пользователя userID=%d: %v", userID, err)
		return nil, err
	}

	logger.Info("Профиль пользователя обновлен: userID=%d", userID)
	return profile, nil
}
