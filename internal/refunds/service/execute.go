package service

import (
	"context"
	"log/slog"

	"github.com/Halturshik/TicketAgregator-API/internal/refunds"
	"github.com/Halturshik/TicketAgregator-API/internal/supplier"
)

func (s *Service) execute(ctx context.Context, operation *refunds.Operation) (*refunds.Result, error) {
	if err := validateProcessingOperation(operation); err != nil {
		return nil, mapErrorContext(ctx, err)
	}
	response, err := s.supplier.ExecuteRefund(ctx, supplierRequest(operation))
	if err != nil {
		operation = s.recordAttemptError(ctx, operation, err)
		slog.WarnContext(ctx, "Не удалось получить от поставщика ответ для возврата",
			slog.Int("refund_id", operation.ID),
			slog.String("supplier_code", operation.SupplierCode),
			slog.Any("error", err),
		)
		return operationResult(operation), nil
	}
	if err := validateSupplierExecution(operation, response); err != nil {
		s.recordAttemptError(ctx, operation, err)
		return nil, mapErrorContext(ctx, err)
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
