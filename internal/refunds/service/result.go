package service

import "github.com/Halturshik/TicketAgregator-API/internal/refunds"

func operationResult(operation *refunds.Operation) *refunds.Result {
	items := make([]refunds.QuoteItem, 0, len(operation.Items))
	for _, item := range operation.Items {
		items = append(items, refunds.QuoteItem{
			TicketID: item.ID, SupplierCode: item.SupplierCode,
			Refundable: operation.Status != refunds.StatusFailed && item.RefundPercent > 0,
			Reason:     item.Reason, RefundPercent: item.RefundPercent,
			GrossAmount:       item.GrossAmount,
			GrossRefundAmount: item.SupplierRefundAmount,
		})
	}
	result := &refunds.Result{
		RefundID: operation.ID, OrderID: operation.Order.ID,
		OrderStatus: operation.Order.Status, Status: operation.Status,
		SupplierRefundID: operation.SupplierRefundID, FailureCode: operation.FailureCode,
		AttemptCount: operation.AttemptCount, CashAmount: operation.CashAmount,
		BonusRestored: operation.BonusRestored, BonusRevoked: operation.BonusRevoked,
		BonusBalance: operation.BonusBalanceAfter, BonusDebt: operation.BonusDebtAfter,
		Items: items,
	}
	if operation.Status == refunds.StatusProcessing {
		nextRetryAt := operation.NextRetryAt
		result.NextRetryAt = &nextRetryAt
	}
	return result
}
