package service

import (
	"context"
	"errors"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/documents"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
)

func (s *Service) Delete(ctx context.Context, ownerUserID int, documentID int) error {
	if documentID <= 0 {
		return apierror.ErrInvalidRequest
	}
	if err := s.repo.Delete(ctx, ownerUserID, documentID); errors.Is(err, documents.ErrNotFound) {
		return apierror.ErrNotFound
	} else if err != nil {
		logger.Error("Ошибка удаления документа userID=%d documentID=%d: %v", ownerUserID, documentID, err)
		return err
	}
	logger.Info("Документ удален: userID=%d documentID=%d", ownerUserID, documentID)
	return nil
}
