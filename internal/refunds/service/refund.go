package service

import (
	"context"
	"errors"

	"github.com/Halturshik/TicketAgregator-API/internal/refunds"
)

func (s *Service) Refund(ctx context.Context, userID *int, orderID int, input refunds.Input) (*refunds.Result, error) {
	if err := validateRefundInput(orderID, input); err != nil {
		return nil, err
	}
	requestHash := refundRequestHash(orderID, input)
	existing, err := s.repo.GetByKey(ctx, input.IdempotencyKey)
	if err == nil {
		return s.resume(ctx, existing, userID, input.GuestPaymentToken, requestHash)
	}
	if !errors.Is(err, refunds.ErrNotFound) {
		return nil, mapErrorContext(ctx, err)
	}

	prepared, err := s.prepare(ctx, userID, orderID, input)
	if err != nil {
		return s.resumeAfterPreparationError(ctx, userID, input, requestHash, err)
	}
	if !prepared.quote.Refundable {
		return nil, mapErrorContext(ctx, refunds.ErrNotAllowed)
	}
	supplierCode, err := quoteSupplier(prepared.quote)
	if err != nil {
		return nil, mapErrorContext(ctx, err)
	}
	reconciliationStartedAt := s.now().UTC()
	created, err := s.createOperation(
		ctx, userID, orderID, input, requestHash,
		supplierCode, reconciliationStartedAt, prepared,
	)
	if err != nil {
		return nil, mapErrorContext(ctx, err)
	}

	operation, err := s.repo.GetByKey(ctx, input.IdempotencyKey)
	if err != nil {
		return nil, mapErrorContext(ctx, err)
	}
	if !created {
		return s.resume(ctx, operation, userID, input.GuestPaymentToken, requestHash)
	}
	return s.execute(ctx, operation)
}

func (s *Service) resumeAfterPreparationError(
	ctx context.Context,
	userID *int,
	input refunds.Input,
	requestHash string,
	preparationErr error,
) (*refunds.Result, error) {
	existing, err := s.repo.GetByKey(ctx, input.IdempotencyKey)
	if err == nil {
		return s.resume(ctx, existing, userID, input.GuestPaymentToken, requestHash)
	}
	if !errors.Is(err, refunds.ErrNotFound) {
		return nil, mapErrorContext(ctx, err)
	}
	return nil, mapErrorContext(ctx, preparationErr)
}

func (s *Service) resume(
	ctx context.Context,
	operation *refunds.Operation,
	userID *int,
	guestToken string,
	requestHash string,
) (*refunds.Result, error) {
	if err := authorize(&operation.Order, userID, guestToken); err != nil {
		return nil, mapErrorContext(ctx, err)
	}
	if operation.RequestHash != requestHash {
		return nil, mapErrorContext(ctx, refunds.ErrIdempotencyConflict)
	}
	if operation.Status != refunds.StatusProcessing {
		return operationResult(operation), nil
	}
	return s.executeIfDue(ctx, operation)
}
