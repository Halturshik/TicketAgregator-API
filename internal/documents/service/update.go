package service

import (
	"context"
	"errors"
	"log/slog"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/documents"
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
		return nil, err
	}
	slog.InfoContext(ctx, "Документ обновлён",
		slog.Int("user_id", ownerUserID),
		slog.Int("document_id", documentID),
		slog.String("verification_status", prepared.status),
	)
	return document, nil
}
