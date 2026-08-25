package service

import (
	"context"
	"errors"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/documents"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
)

func (s *Service) Update(ctx context.Context, ownerUserID int, documentID int, in documents.SaveDocumentInput) (*documents.Document, error) {
	if documentID <= 0 {
		return nil, apierror.ErrInvalidRequest
	}
	prepared, err := s.prepareDocument(&in)
	if err != nil {
		return nil, err
	}
	document, err := s.repo.Update(
		ctx, ownerUserID, documentID, in,
		prepared.status, prepared.fingerprint, prepared.expiresAt, prepared.checkedAt,
	)
	if errors.Is(err, documents.ErrNotFound) {
		return nil, apierror.ErrNotFound
	}
	if errors.Is(err, documents.ErrAlreadyExists) {
		return nil, apierror.ErrDocumentAlreadyExists
	}
	if err != nil {
		logger.Error("Ошибка обновления документа userID=%d documentID=%d: %v", ownerUserID, documentID, err)
		return nil, err
	}
	logger.Info("Документ обновлен: userID=%d documentID=%d status=%s", ownerUserID, documentID, prepared.status)
	return document, nil
}
