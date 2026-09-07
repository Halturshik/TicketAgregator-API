package service

import (
	"context"
	"errors"
	"strings"

	"github.com/Halturshik/TicketAgregator-API/internal/bookingaccess"
	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
)

const accessRequestMessage = "Если email совпадает с данными заказа, код подтверждения отправлен"

func (s *Service) RequestAccess(
	ctx context.Context,
	input bookingaccess.AccessRequestInput,
) (*bookingaccess.AccessRequestOutput, error) {
	location, err := locator(input.TicketNumber, input.OrderNumber)
	if err != nil {
		return nil, err
	}
	email, err := validEmail(input.Email)
	if err != nil {
		return nil, err
	}
	challengeID := s.newChallengeID()
	output := &bookingaccess.AccessRequestOutput{
		ChallengeID: challengeID,
		Message:     accessRequestMessage,
	}

	code, err := s.codes.GenerateVerificationCode()
	if err != nil {
		return nil, mapError("генерации кода доступа к заказу", err)
	}
	locationValue := location.OrderNumber
	if locationValue == "" {
		locationValue = location.TicketNumber
	}
	booking, err := s.repo.GuestByLocator(ctx, location, email)
	if errors.Is(err, bookingaccess.ErrNotFound) {
		dummy := bookingaccess.Challenge{
			ID: challengeID, Locator: locationValue, Email: email,
		}
		if err := s.challenges.Request(ctx, dummy, code); err != nil {
			return nil, mapError("ограничения запроса доступа к заказу", err)
		}
		logger.Warn("Запрошен доступ к гостевому заказу с несовпадающими данными")
		return output, nil
	}
	if err != nil {
		return nil, mapError("запроса доступа к гостевому заказу", err)
	}
	challenge := bookingaccess.Challenge{
		ID: challengeID, OrderID: booking.ID,
		OrderNumber: booking.OrderNumber, Locator: locationValue, Email: booking.GuestEmail,
	}
	if err := s.challenges.Request(ctx, challenge, code); err != nil {
		return nil, mapError("сохранения кода доступа к заказу", err)
	}
	if err := s.mailer.SendVerificationEmail(ctx, booking.GuestEmail, code); err != nil {
		return nil, mapError("отправки кода доступа к заказу", err)
	}
	logger.Info("Отправлен код доступа к гостевому заказу order_id=%d", booking.ID)
	return output, nil
}

func (s *Service) ConfirmAccess(
	ctx context.Context,
	input bookingaccess.AccessConfirmInput,
) (*bookingaccess.AccessOutput, error) {
	input.ChallengeID = strings.TrimSpace(input.ChallengeID)
	input.Code = strings.TrimSpace(input.Code)
	if err := validConfirmation(input); err != nil {
		return nil, err
	}
	challenge, err := s.challenges.Verify(ctx, input.ChallengeID, input.Code)
	if err != nil {
		return nil, mapError("проверки кода доступа к заказу", err)
	}
	if challenge.OrderID <= 0 || challenge.OrderNumber == "" {
		return nil, apierror.ErrCodeExpired
	}
	token, err := s.access.Issue(ctx, challenge.OrderID)
	if err != nil {
		return nil, mapError("выдачи токена доступа к заказу", err)
	}
	logger.Info("Выдан токен доступа к гостевому заказу order_id=%d", challenge.OrderID)
	return &bookingaccess.AccessOutput{
		OrderNumber: challenge.OrderNumber,
		ExpiresAt:   s.now().UTC().Add(bookingaccess.AccessTokenTTL),
		Token:       token,
	}, nil
}
