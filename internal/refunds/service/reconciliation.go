package service

import (
	"context"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
	"github.com/Halturshik/TicketAgregator-API/internal/refunds"
)

const (
	reconcileDefaultLimit = 20
	reconcileMaxLimit     = 100
	attemptRecordTimeout  = 3 * time.Second
)

func (s *Service) Reconcile(ctx context.Context, limit int) error {
	limit = normalizeReconciliationLimit(limit)
	now := s.now().UTC()
	reviewed, err := s.repo.MarkReviewDue(ctx, now, refunds.MaxReconciliationAttempts)
	if err != nil {
		return err
	}
	if reviewed > 0 {
		logger.Warn("Автоматическая обработка возвратов остановлена: requiresReview=%d", reviewed)
	}
	operations, err := s.repo.ListProcessing(ctx, now, refunds.MaxReconciliationAttempts, limit)
	if err != nil {
		return err
	}
	for index := range operations {
		if _, err := s.execute(ctx, &operations[index]); err != nil {
			logger.Error("Ошибка повторной обработки возврата refundID=%d: %v", operations[index].ID, err)
		}
	}
	return nil
}

func (s *Service) executeIfDue(ctx context.Context, operation *refunds.Operation) (*refunds.Result, error) {
	now := s.now().UTC()
	if operation.AttemptCount >= refunds.MaxReconciliationAttempts ||
		!operation.ReconciliationDeadline.After(now) {
		if _, err := s.repo.MarkReviewDue(ctx, now, refunds.MaxReconciliationAttempts); err != nil {
			return nil, mapError(err)
		}
		latest, err := s.repo.GetByKey(ctx, operation.IdempotencyKey)
		if err != nil {
			return nil, mapError(err)
		}
		return operationResult(latest), nil
	}
	if operation.NextRetryAt.After(now) {
		return operationResult(operation), nil
	}
	return s.execute(ctx, operation)
}

func (s *Service) recordAttemptError(operation *refunds.Operation, cause error) *refunds.Operation {
	now := s.now().UTC()
	failedAttempts := operation.AttemptCount + 1
	nextRetryAt := now.Add(refunds.ReconciliationRetryDelay(failedAttempts))
	status := refunds.StatusProcessing
	failureCode := ""
	if failedAttempts >= refunds.MaxReconciliationAttempts ||
		!operation.ReconciliationDeadline.After(now) ||
		!nextRetryAt.Before(operation.ReconciliationDeadline) {
		status = refunds.StatusRequiresReview
		failureCode = refunds.FailureReconciliationExhausted
		nextRetryAt = operation.ReconciliationDeadline
	}

	ctx, cancel := context.WithTimeout(context.Background(), attemptRecordTimeout)
	defer cancel()
	if err := s.repo.RecordAttemptError(ctx, refunds.AttemptErrorParams{
		RefundID: operation.ID, Message: cause.Error(), Status: status,
		FailureCode: failureCode, NextRetryAt: nextRetryAt,
	}); err != nil {
		logger.Error("Не удалось сохранить ошибку попытки возврата refundID=%d: %v", operation.ID, err)
		return operation
	}
	latest, err := s.repo.GetByKey(ctx, operation.IdempotencyKey)
	if err != nil {
		logger.Error("Не удалось перечитать возврат после ошибки refundID=%d: %v", operation.ID, err)
		return operation
	}
	if latest.Status == refunds.StatusRequiresReview {
		logger.Warn(
			"Автоматическая обработка возврата исчерпана: refundID=%d attempts=%d",
			latest.ID, latest.AttemptCount,
		)
	}
	return latest
}

func normalizeReconciliationLimit(limit int) int {
	if limit <= 0 {
		return reconcileDefaultLimit
	}
	if limit > reconcileMaxLimit {
		return reconcileMaxLimit
	}
	return limit
}
