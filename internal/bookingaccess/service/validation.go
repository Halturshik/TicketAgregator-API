package service

import (
	"strings"

	"github.com/Halturshik/TicketAgregator-API/internal/bookingaccess"
	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/common/cleaning"
	"github.com/Halturshik/TicketAgregator-API/internal/common/validator"
	ordernumber "github.com/Halturshik/TicketAgregator-API/internal/orders/number"
	"github.com/google/uuid"
)

func locator(ticketNumber string, orderNumber string) (bookingaccess.Locator, error) {
	ticketNumber = normalizeNumber(ticketNumber)
	orderNumber = normalizeNumber(orderNumber)
	if (ticketNumber == "") == (orderNumber == "") {
		return bookingaccess.Locator{}, apierror.ErrInvalidRequest
	}
	if ticketNumber != "" {
		if !ordernumber.ValidTicket(ticketNumber) {
			return bookingaccess.Locator{}, apierror.ErrInvalidRequest
		}
		return bookingaccess.Locator{TicketNumber: ticketNumber}, nil
	}
	if !ordernumber.ValidOrder(orderNumber) {
		return bookingaccess.Locator{}, apierror.ErrInvalidRequest
	}
	return bookingaccess.Locator{OrderNumber: orderNumber}, nil
}

func validOrderNumber(value string) (string, error) {
	value = normalizeNumber(value)
	if !ordernumber.ValidOrder(value) {
		return "", apierror.ErrInvalidRequest
	}
	return value, nil
}

func validEmail(value string) (string, error) {
	value = cleaning.Email(value)
	if !validator.ValidEmail(value) {
		return "", apierror.Validation(map[string]string{apierror.FieldEmail: apierror.ErrInvalidEmail})
	}
	return value, nil
}

func validConfirmation(input bookingaccess.AccessConfirmInput) error {
	input.ChallengeID = strings.TrimSpace(input.ChallengeID)
	input.Code = strings.TrimSpace(input.Code)
	if uuid.Validate(input.ChallengeID) != nil || input.Code == "" {
		return apierror.ErrInvalidRequest
	}
	return nil
}

func normalizeNumber(value string) string {
	return strings.ToUpper(strings.Join(strings.Fields(value), ""))
}
