package service

import (
	"context"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/common/validator"
	"github.com/Halturshik/TicketAgregator-API/internal/documents"
)

func (s *Service) ValidateForBooking(ctx context.Context, in documents.BookingValidationInput) (*documents.ValidatedBookingDocument, error) {
	if in.DepartureTime.IsZero() || in.Transport == "" {
		return nil, apierror.ErrInvalidRequest
	}
	birthDate, err := validateBookingPassenger(in)
	if err != nil {
		return nil, err
	}
	document, err := s.resolveBookingDocument(ctx, in)
	if err != nil {
		return nil, err
	}
	if !validator.ValidDocumentNumber(document.Type, document.Number) {
		return nil, apierror.ErrInvalidDocument
	}
	if err := s.validateDocumentRule(ctx, in, document.Type, birthDate); err != nil {
		return nil, err
	}
	if err := validateDocumentExpiration(document.ExpiresAt, in.DepartureTime); err != nil {
		return nil, err
	}
	return s.verifyBookingDocument(ctx, document)
}
