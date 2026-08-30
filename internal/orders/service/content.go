package service

import (
	"context"
	"strings"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/documents"
	"github.com/Halturshik/TicketAgregator-API/internal/orders"
	"github.com/Halturshik/TicketAgregator-API/internal/passengers"
	"github.com/Halturshik/TicketAgregator-API/internal/search"
)

func (s *Service) buildOrderContent(
	ctx context.Context,
	userID *int,
	trip *selectedTrip,
	bookings []orders.PassengerBooking,
) ([]orders.OrderPassengerDraft, []orders.TicketDraft, error) {
	passengerDrafts := make([]orders.OrderPassengerDraft, 0, len(bookings))
	seenSavedPassengers := make(map[int]struct{})
	selfSeen := false

	for _, booking := range bookings {
		if strings.EqualFold(strings.TrimSpace(booking.Source), passengers.SourceSelf) {
			if selfSeen {
				return nil, nil, apierror.ErrInvalidRequest
			}
			selfSeen = true
		}
		draft, err := s.buildPassengerDraft(ctx, userID, trip.directions, booking, seenSavedPassengers)
		if err != nil {
			return nil, nil, err
		}
		passengerDrafts = append(passengerDrafts, *draft)
	}

	tickets, err := buildTickets(trip, passengerDrafts)
	if err != nil {
		return nil, nil, err
	}
	return passengerDrafts, tickets, nil
}

func (s *Service) buildPassengerDraft(
	ctx context.Context,
	userID *int,
	directions []search.Offer,
	booking orders.PassengerBooking,
	seenSavedPassengers map[int]struct{},
) (*orders.OrderPassengerDraft, error) {
	resolved, err := s.resolvePassenger(ctx, userID, booking, seenSavedPassengers)
	if err != nil {
		return nil, err
	}
	var validated *documents.ValidatedBookingDocument
	for _, direction := range directions {
		departure, err := offerDeparture(direction)
		if err != nil {
			return nil, err
		}
		validated, err = s.documents.ValidateForBooking(ctx, bookingDocumentInput(userID, resolved, booking, direction, departure))
		if err != nil {
			return nil, err
		}
	}

	return &orders.OrderPassengerDraft{
		Source:           resolved.Source,
		SavedPassengerID: resolved.SavedPassengerID,
		SavedDocumentID:  validated.SavedDocumentID,
		SaveChanges:      resolved.SaveChanges,
		Passenger:        resolved.Passenger,
		Document: orders.DocumentSnapshot{
			Type: validated.Type, Number: validated.Number, ExpiresAt: validated.ExpiresAt,
			VerificationStatus: validated.VerificationStatus, LastCheckedAt: validated.LastCheckedAt,
		},
		DocumentFingerprint: validated.Fingerprint,
	}, nil
}
