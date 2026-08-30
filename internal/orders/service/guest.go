package service

import (
	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/common/cleaning"
	"github.com/Halturshik/TicketAgregator-API/internal/common/validator"
	"github.com/google/uuid"
)

func guestPaymentData(userID *int, email string) (string, string, error) {
	if userID != nil {
		return "", "", nil
	}
	email = cleaning.Email(email)
	if !validator.ValidEmail(email) {
		return "", "", apierror.ErrInvalidRequest
	}
	return email, uuid.NewString(), nil
}
