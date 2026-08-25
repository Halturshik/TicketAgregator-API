package service

import (
	"net/mail"
	"strings"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/google/uuid"
)

func guestPaymentData(userID *int, email string) (string, string, error) {
	if userID != nil {
		return "", "", nil
	}
	email = strings.TrimSpace(email)
	if _, err := mail.ParseAddress(email); err != nil {
		return "", "", apierror.ErrInvalidRequest
	}
	return email, uuid.NewString(), nil
}
