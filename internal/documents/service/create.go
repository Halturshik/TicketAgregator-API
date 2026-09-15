package service

import (
	"context"
	"errors"
	"log/slog"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/documents"
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
		return nil, err
	}
	slog.InfoContext(ctx, "Документ сохранён",
		slog.Int("user_id", ownerUserID),
		slog.Int("document_id", document.ID),
		slog.String("verification_status", document.VerificationStatus),
	)
	return document, nil
}
