package service

import (
	"context"
	"errors"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/common/validator"
	"github.com/Halturshik/TicketAgregator-API/internal/documents"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
)

func (s *Service) validateDocumentRule(
	ctx context.Context,
	in documents.BookingValidationInput,
	documentType string,
	birthDate time.Time,
) error {
	age := validator.AgeOn(birthDate, in.DepartureTime)
	rule, err := s.repo.FindRule(ctx, in.Transport, in.IsInternational, age, in.Passenger.IsRussian)
	if errors.Is(err, documents.ErrRuleNotFound) {
		logger.Warn("Не найдено правило документа transport=%s international=%t russian=%t", in.Transport, in.IsInternational, in.Passenger.IsRussian)
		return apierror.ErrDocumentNotAllowed
	}
	if err != nil {
		logger.Error("Ошибка поиска правила документа: %v", err)
		return err
	}
	if !rule.Allows(documentType) {
		return apierror.ErrDocumentNotAllowed
	}
	return nil
}
