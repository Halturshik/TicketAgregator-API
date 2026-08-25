package service

import (
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/documents"
	"github.com/Halturshik/TicketAgregator-API/internal/orders"
	"github.com/Halturshik/TicketAgregator-API/internal/search"
)

func bookingDocumentInput(
	userID *int,
	resolved *resolvedPassenger,
	booking orders.PassengerBooking,
	direction search.Offer,
	departure time.Time,
) documents.BookingValidationInput {
	return documents.BookingValidationInput{
		OwnerUserID:      userID,
		SavedPassengerID: resolved.SavedPassengerID,
		Passenger: documents.BookingPassenger{
			FirstName: resolved.Passenger.FirstName, MiddleName: resolved.Passenger.MiddleName,
			LastName: resolved.Passenger.LastName, BirthDate: resolved.Passenger.BirthDate,
			IsRussian: resolved.Passenger.IsRussian,
		},
		Document: documents.BookingDocument{
			ID: booking.Document.SavedDocumentID, Type: booking.Document.Type,
			Number: booking.Document.Number, ExpiresAt: booking.Document.ExpiresAt,
		},
		Transport: direction.Transport, IsInternational: direction.IsInternational, DepartureTime: departure,
	}
}
