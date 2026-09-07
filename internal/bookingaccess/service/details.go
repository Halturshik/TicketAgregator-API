package service

import (
	"context"

	"github.com/Halturshik/TicketAgregator-API/internal/bookingaccess"
)

func (s *Service) Details(
	ctx context.Context,
	userID *int,
	orderNumber string,
	accessToken string,
) (*bookingaccess.BookingDetails, error) {
	orderNumber, err := validOrderNumber(orderNumber)
	if err != nil {
		return nil, err
	}
	booking, err := s.repo.ByOrderNumber(ctx, orderNumber)
	if err != nil {
		return nil, mapError("получения бронирования", err)
	}
	if err := s.authorize(ctx, booking, userID, accessToken); err != nil {
		return nil, mapError("проверки доступа к бронированию", err)
	}
	return bookingDetails(booking), nil
}

func (s *Service) authorize(
	ctx context.Context,
	booking *bookingaccess.Booking,
	userID *int,
	accessToken string,
) error {
	if booking.UserID != nil {
		if userID == nil || *userID != *booking.UserID {
			return bookingaccess.ErrForbidden
		}
		return nil
	}
	return s.access.Authorize(ctx, accessToken, booking.ID)
}
