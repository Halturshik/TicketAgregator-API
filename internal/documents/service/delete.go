package service

import (
	"context"
	"errors"
	"log/slog"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/documents"
)

func (s *Service) Delete(ctx context.Context, ownerUserID int, documentID int) error {
	if documentID <= 0 {
		return apierror.ErrInvalidRequest
	}
	if err := s.repo.Delete(ctx, ownerUserID, documentID); errors.Is(err, documents.ErrNotFound) {
		return apierror.ErrNotFound
	} else if err != nil {
		return err
	}
	slog.InfoContext(ctx, "Документ удалён",
		slog.Int("user_id", ownerUserID),
		slog.Int("document_id", documentID),
	)
	return nil
}
