package repository

import (
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/documents"
	"github.com/Halturshik/TicketAgregator-API/internal/passengers"
)

func scanPassenger(scanner interface {
	Scan(dest ...any) error
}) (*passengers.Passenger, error) {
	var passenger passengers.Passenger
	var birthDate time.Time
	if err := scanner.Scan(
		&passenger.ID, &passenger.FirstName, &passenger.MiddleName,
		&passenger.LastName, &birthDate, &passenger.IsRussian,
	); err != nil {
		return nil, err
	}
	passenger.BirthDate = birthDate.Format(documents.DateLayout)
	return &passenger, nil
}
