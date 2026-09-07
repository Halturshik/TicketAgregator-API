package service

import (
	"errors"

	"github.com/Halturshik/TicketAgregator-API/internal/bookingaccess"
	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
)

func mapError(operation string, err error) error {
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
		logger.Error("Ошибка %s: %v", operation, err)
		return err
	}
}
