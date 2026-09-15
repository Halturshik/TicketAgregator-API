package service

import (
	"context"
	"log/slog"
	"time"

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
		slog.WarnContext(ctx, "Автоматическая обработка незавершённых возвратов остановлена после исчерпания лимита",
			slog.Int("requires_review", reviewed),
		)
	}
	operations, err := s.repo.ListProcessing(ctx, now, refunds.MaxReconciliationAttempts, limit)
	if err != nil {
		return err
	}
	for index := range operations {
		if _, err := s.execute(ctx, &operations[index]); err != nil {
			slog.ErrorContext(ctx, "Ошибка при повторной обработке незавершённого возврата",
				slog.Int("refund_id", operations[index].ID),
				slog.Any("error", err),
			)
		}
	}
	return nil
}

func (s *Service) executeIfDue(ctx context.Context, operation *refunds.Operation) (*refunds.Result, error) {
	now := s.now().UTC()
	if operation.AttemptCount >= refunds.MaxReconciliationAttempts ||
		!operation.ReconciliationDeadline.After(now) {
		if _, err := s.repo.MarkReviewDue(ctx, now, refunds.MaxReconciliationAttempts); err != nil {
			return nil, mapErrorContext(ctx, err)
		}
		latest, err := s.repo.GetByKey(ctx, operation.IdempotencyKey)
		if err != nil {
			return nil, mapErrorContext(ctx, err)
		}
		return operationResult(latest), nil
	}
	if operation.NextRetryAt.After(now) {
		return operationResult(operation), nil
	}
	return s.execute(ctx, operation)
}

func (s *Service) recordAttemptError(parent context.Context, operation *refunds.Operation, cause error) *refunds.Operation {
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

	ctx, cancel := context.WithTimeout(context.WithoutCancel(parent), attemptRecordTimeout)
	defer cancel()
	if err := s.repo.RecordAttemptError(ctx, refunds.AttemptErrorParams{
		RefundID: operation.ID, Message: cause.Error(), Status: status,
		FailureCode: failureCode, NextRetryAt: nextRetryAt,
	}); err != nil {
		slog.ErrorContext(ctx, "Не удалось сохранить ошибку, полученную при попытке возврата",
			slog.Int("refund_id", operation.ID),
			slog.Any("error", err),
		)
		return operation
	}
	latest, err := s.repo.GetByKey(ctx, operation.IdempotencyKey)
	if err != nil {
		slog.ErrorContext(ctx, "Ошибка при чтении актуального состояния неуспешной попытки возврата",
			slog.Int("refund_id", operation.ID),
			slog.Any("error", err),
		)
		return operation
	}
	if latest.Status == refunds.StatusRequiresReview {
		slog.WarnContext(ctx, "Исчерпаны попытки автоматической обработки возврата",
			slog.Int("refund_id", latest.ID),
			slog.Int("attempts", latest.AttemptCount),
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
