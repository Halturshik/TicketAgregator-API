package service

import (
	"strings"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/common/validator"
	"github.com/Halturshik/TicketAgregator-API/internal/documents"
)

func validateBookingPassenger(in documents.BookingValidationInput) (time.Time, error) {
	useLatinNames := !in.Passenger.IsRussian || in.IsInternational
	if !validBookingNames(in.Passenger, useLatinNames) {
		return time.Time{}, apierror.Validation(map[string]string{
			apierror.FieldPassengers: "Некорректно указаны данные пассажира",
		})
	}
	birthDate, ok := validator.ValidPassengerBirthDate(in.Passenger.BirthDate, in.DepartureTime)
	if !ok {
		return time.Time{}, apierror.Validation(map[string]string{
			apierror.FieldBirthDate: "Некорректно указана дата рождения пассажира",
		})
	}
	return birthDate, nil
}

func validBookingNames(passenger documents.BookingPassenger, useLatinNames bool) bool {
	validName := validator.ValidCyrillicName
	if useLatinNames {
		validName = validator.ValidLatinName
	}
	if !validName(passenger.FirstName) || !validName(passenger.LastName) {
		return false
	}
	return strings.TrimSpace(passenger.MiddleName) == "" || validName(passenger.MiddleName)
}
