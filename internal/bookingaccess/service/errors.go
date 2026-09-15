package service

import (
	"errors"

	"github.com/Halturshik/TicketAgregator-API/internal/bookingaccess"
	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
)

func mapError(_ string, err error) error {
	var apiErr *apierror.APIError
	if errors.As(err, &apiErr) {
		return apiErr
	}
	switch {
	case errors.Is(err, bookingaccess.ErrNotFound):
		return apierror.ErrNotFound
	case errors.Is(err, bookingaccess.ErrForbidden):
		return apierror.ErrNotFound
	case errors.Is(err, bookingaccess.ErrInvalidToken):
		return apierror.ErrInvalidToken
	default:
		return err
	}
}
