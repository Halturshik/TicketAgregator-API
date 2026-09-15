package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/common/validator"
	"github.com/Halturshik/TicketAgregator-API/internal/documents"
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
		slog.WarnContext(ctx, "Не найдено правило документа",
			slog.String("transport", in.Transport),
			slog.Bool("international", in.IsInternational),
			slog.Bool("russian_citizen", in.Passenger.IsRussian),
		)
		return apierror.ErrDocumentNotAllowed
	}
	if err != nil {
		return err
	}
	if !rule.Allows(documentType) {
		return apierror.ErrDocumentNotAllowed
	}
	return nil
}
