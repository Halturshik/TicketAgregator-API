package service

import (
	"context"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/documents"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
)

func validateDocumentExpiration(expiresAt string, departure time.Time) error {
	if expiresAt == "" {
		return nil
	}
	parsed, err := time.Parse(documents.DateLayout, expiresAt)
	if err != nil {
		return apierror.ErrInvalidDocument
	}
	if parsed.Before(dateOnly(departure)) {
		return apierror.ErrDocumentExpired
	}
	return nil
}

func (s *Service) verifyBookingDocument(
	ctx context.Context,
	document *documents.ValidatedBookingDocument,
) (*documents.ValidatedBookingDocument, error) {
	checkedAt := s.now().UTC()
	status, fingerprint := verifyDocument(s.verificationSecret, document.Type, document.Number, checkedAt)
	if document.SavedDocumentID != nil {
		if err := s.repo.UpdateVerification(ctx, *document.SavedDocumentID, status, checkedAt); err != nil {
			logger.Error("Ошибка сохранения результата повторной проверки documentID=%d: %v", *document.SavedDocumentID, err)
			return nil, err
		}
	}
	document.VerificationStatus = status
	document.LastCheckedAt = checkedAt.Format(time.RFC3339)
	document.Fingerprint = fingerprint
	if status == documents.StatusRejected {
		return nil, apierror.ErrDocumentRejected
	}
	return document, nil
}

func dateOnly(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}
