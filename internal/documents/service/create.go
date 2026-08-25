package service

import (
	"context"
	"errors"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/documents"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
)

func (s *Service) Create(ctx context.Context, ownerUserID int, in documents.SaveDocumentInput) (*documents.Document, error) {
	prepared, err := s.prepareDocument(&in)
	if err != nil {
		return nil, err
	}

	var document *documents.Document
	if in.PassengerID != nil {
		document, err = s.repo.CreateForPassenger(
			ctx, ownerUserID, *in.PassengerID, in,
			prepared.status, prepared.fingerprint, prepared.expiresAt, prepared.checkedAt,
		)
	} else {
		document, err = s.repo.CreateForUser(
			ctx, ownerUserID, in,
			prepared.status, prepared.fingerprint, prepared.expiresAt, prepared.checkedAt,
		)
	}
	if errors.Is(err, documents.ErrNotFound) {
		return nil, apierror.ErrNotFound
	}
	if errors.Is(err, documents.ErrAlreadyExists) {
		return nil, apierror.ErrDocumentAlreadyExists
	}
	if err != nil {
		logger.Error("Ошибка сохранения документа userID=%d type=%s: %v", ownerUserID, in.Type, err)
		return nil, err
	}
	logger.Info("Документ сохранен: userID=%d documentID=%d status=%s", ownerUserID, document.ID, document.VerificationStatus)
	return document, nil
}
