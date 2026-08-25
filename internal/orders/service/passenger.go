package service

import (
	"context"
	"strings"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/orders"
	"github.com/Halturshik/TicketAgregator-API/internal/passengers"
)

type resolvedPassenger struct {
	Source           string
	SavedPassengerID *int
	SaveChanges      bool
	Passenger        orders.PassengerSnapshot
}

func (s *Service) resolvePassenger(
	ctx context.Context,
	userID *int,
	booking orders.PassengerBooking,
	seenSavedPassengers map[int]struct{},
) (*resolvedPassenger, error) {
	source := strings.ToLower(strings.TrimSpace(booking.Source))
	switch source {
	case passengers.SourceNew:
		return resolveNewPassenger(booking, source)
	case passengers.SourceSelf:
		return selfPassenger(userID, booking, source)
	case passengers.SourceSaved:
		return s.savedPassenger(ctx, userID, booking, source, seenSavedPassengers)
	default:
		return nil, apierror.ErrInvalidRequest
	}
}

func resolveNewPassenger(booking orders.PassengerBooking, source string) (*resolvedPassenger, error) {
	if booking.SavedPassengerID != nil || booking.Document.SavedDocumentID != nil {
		return nil, apierror.ErrInvalidRequest
	}
	return &resolvedPassenger{Source: source, Passenger: booking.Passenger}, nil
}

func selfPassenger(userID *int, booking orders.PassengerBooking, source string) (*resolvedPassenger, error) {
	if userID == nil {
		return nil, apierror.ErrUnauthorized
	}
	if booking.SavedPassengerID != nil || booking.SaveChanges {
		return nil, apierror.ErrInvalidRequest
	}
	return &resolvedPassenger{Source: source, Passenger: booking.Passenger}, nil
}

func (s *Service) savedPassenger(
	ctx context.Context,
	userID *int,
	booking orders.PassengerBooking,
	source string,
	seen map[int]struct{},
) (*resolvedPassenger, error) {
	if userID == nil {
		return nil, apierror.ErrUnauthorized
	}
	if booking.SavedPassengerID == nil || *booking.SavedPassengerID <= 0 {
		return nil, apierror.ErrInvalidRequest
	}
	passengerID := *booking.SavedPassengerID
	if _, exists := seen[passengerID]; exists {
		return nil, apierror.ErrInvalidRequest
	}
	saved, err := s.passengers.GetOwned(ctx, *userID, passengerID)
	if err != nil {
		return nil, err
	}
	seen[passengerID] = struct{}{}

	passenger := booking.Passenger
	if emptyPassengerSnapshot(passenger) {
		passenger = orders.PassengerSnapshot{
			FirstName: saved.FirstName, MiddleName: saved.MiddleName, LastName: saved.LastName,
			BirthDate: saved.BirthDate, IsRussian: saved.IsRussian,
		}
	}
	return &resolvedPassenger{
		Source: source, SavedPassengerID: &passengerID, SaveChanges: booking.SaveChanges, Passenger: passenger,
	}, nil
}

func emptyPassengerSnapshot(passenger orders.PassengerSnapshot) bool {
	return strings.TrimSpace(passenger.FirstName) == "" &&
		strings.TrimSpace(passenger.MiddleName) == "" &&
		strings.TrimSpace(passenger.LastName) == "" &&
		strings.TrimSpace(passenger.BirthDate) == ""
}
