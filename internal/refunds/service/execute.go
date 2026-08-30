package service

import (
	"context"

	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
	"github.com/Halturshik/TicketAgregator-API/internal/refunds"
	"github.com/Halturshik/TicketAgregator-API/internal/supplier"
)

func (s *Service) execute(ctx context.Context, operation *refunds.Operation) (*refunds.Result, error) {
	if err := validateProcessingOperation(operation); err != nil {
		return nil, mapError(err)
	}
	response, err := s.supplier.ExecuteRefund(ctx, supplierRequest(operation))
	if err != nil {
		operation = s.recordAttemptError(operation, err)
		logger.Warn("Поставщик не подтвердил возврат: refundID=%d supplier=%s: %v", operation.ID, operation.SupplierCode, err)
		return operationResult(operation), nil
	}
	if err := validateSupplierExecution(operation, response); err != nil {
		s.recordAttemptError(operation, err)
		return nil, mapError(err)
	}
	if response.Status == supplier.RefundStatusRejected {
		return s.finalizeFailure(ctx, operation, response)
	}
	return s.finalizeSuccess(ctx, operation, response)
}

func supplierRequest(operation *refunds.Operation) supplier.ExecuteRefundRequest {
	items := make([]supplier.ExecuteRefundItem, 0, len(operation.Items))
	for _, item := range operation.Items {
		items = append(items, supplier.ExecuteRefundItem{
			TicketID: item.ID, TicketNumber: item.TicketNumber,
			SupplierOfferID: item.SupplierOfferID, FareType: item.FareType,
			DepartureUnix: item.DepartureAt.Unix(), GrossAmount: item.GrossAmount,
		})
	}
	return supplier.ExecuteRefundRequest{
		ProviderCode: operation.SupplierCode, IdempotencyKey: operation.IdempotencyKey, Items: items,
	}
}
