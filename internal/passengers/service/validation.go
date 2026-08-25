package service

import (
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/common/cleaning"
	"github.com/Halturshik/TicketAgregator-API/internal/common/validator"
	"github.com/Halturshik/TicketAgregator-API/internal/passengers"
)

func (s *Service) validate(in *passengers.SavePassengerInput) (time.Time, error) {
	in.FirstName = cleaning.Name(in.FirstName)
	in.MiddleName = cleaning.Name(in.MiddleName)
	in.LastName = cleaning.Name(in.LastName)

	fields := map[string]string{}
	if !validator.NotEmpty(in.FirstName) || !validSavedPassengerName(in.FirstName) {
		fields[apierror.FieldFirstName] = apierror.ErrInvalidFirstName
	}
	if validator.NotEmpty(in.MiddleName) && !validSavedPassengerName(in.MiddleName) {
		fields[apierror.FieldMiddleName] = apierror.ErrInvalidMiddleName
	}
	if !validator.NotEmpty(in.LastName) || !validSavedPassengerName(in.LastName) {
		fields[apierror.FieldLastName] = apierror.ErrInvalidLastName
	}
	parsedBirthDate, parseErr := validPassengerBirthDate(in.BirthDate)
	if parseErr != nil {
		fields[apierror.FieldBirthDate] = apierror.ErrInvalidBirthDate
	}
	if len(fields) > 0 {
		return time.Time{}, apierror.Validation(fields)
	}
	return parsedBirthDate, nil
}

func validSavedPassengerName(value string) bool {
	return validator.ValidCyrillicName(value) || validator.ValidLatinName(value)
}

func validPassengerBirthDate(value string) (time.Time, error) {
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return time.Time{}, err
	}
	if parsed.After(time.Now()) {
		return time.Time{}, apierror.ErrInvalidRequest
	}
	return parsed, nil
}
