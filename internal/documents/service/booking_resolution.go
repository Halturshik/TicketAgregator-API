package service

import (
	"context"
	"errors"
	"strings"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/documents"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
)

func (s *Service) resolveBookingDocument(ctx context.Context, in documents.BookingValidationInput) (*documents.ValidatedBookingDocument, error) {
	if in.Document.ID == nil {
		return &documents.ValidatedBookingDocument{
			Type:      strings.TrimSpace(in.Document.Type),
			Number:    strings.ToUpper(strings.TrimSpace(in.Document.Number)),
			ExpiresAt: strings.TrimSpace(in.Document.ExpiresAt),
		}, nil
	}
	if in.OwnerUserID == nil {
		return nil, apierror.ErrForbidden
	}
	document, err := s.repo.GetOwned(ctx, *in.OwnerUserID, *in.Document.ID, in.SavedPassengerID)
	if errors.Is(err, documents.ErrNotFound) {
		return nil, apierror.ErrNotFound
	}
	if err != nil {
		logger.Error("Ошибка проверки принадлежности документа userID=%d documentID=%d: %v", *in.OwnerUserID, *in.Document.ID, err)
		return nil, err
	}
	return &documents.ValidatedBookingDocument{
		SavedDocumentID: &document.ID,
		Type:            document.Type,
		Number:          document.Number,
		ExpiresAt:       document.ExpiresAt,
	}, nil
}
