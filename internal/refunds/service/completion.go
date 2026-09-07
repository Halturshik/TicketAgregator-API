package service

import (
	"context"

	"github.com/Halturshik/TicketAgregator-API/internal/orders"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
	"github.com/Halturshik/TicketAgregator-API/internal/refunds"
	"github.com/Halturshik/TicketAgregator-API/internal/supplier"
)

func (s *Service) finalizeSuccess(
	ctx context.Context,
	operation *refunds.Operation,
	response *supplier.ExecuteRefundResult,
) (*refunds.Result, error) {
	responseByID := make(map[int]supplier.ExecuteRefundItemResult, len(response.Items))
	for _, item := range response.Items {
		responseByID[item.TicketID] = item
	}
	err := s.repo.WithinTransaction(ctx, func(tx refunds.Transaction) error {
		order, err := tx.LockOrder(ctx, operation.Order.ID)
		if err != nil {
			return err
		}
		current, err := tx.LockOperation(ctx, operation.ID)
		if err != nil {
			return err
		}
		if current.Status != refunds.StatusProcessing {
			return nil
		}
		if err := validateProcessingOperation(current); err != nil {
			return err
		}
		if err := validateSupplierExecution(current, response); err != nil {
			return err
		}

		success := refunds.SuccessParams{
			RefundID: operation.ID, SupplierRefundID: response.SupplierRefundID,
			Items: make([]refunds.SuccessItemParams, 0, len(current.Items)),
		}
		ids := make([]int, 0, len(current.Items))
		refundedTicketTotal := 0
		supplierRefundAmount := 0
		for _, item := range current.Items {
			supplierItem := responseByID[item.ID]
			refundedTicketTotal += item.GrossAmount
			supplierRefundAmount += supplierItem.RefundAmount
			success.Items = append(success.Items, refunds.SuccessItemParams{
				TicketID: item.ID, Reason: supplierItem.Reason,
				RefundPercent:        supplierItem.RefundPercent,
				SupplierRefundAmount: supplierItem.RefundAmount,
			})
			ids = append(ids, item.ID)
		}
		financials, err := calculateRefundFinancials(order, refundedTicketTotal, supplierRefundAmount)
		if err != nil {
			return err
		}
		success.CashAmount = financials.cashAmount
		success.BonusRestored = financials.bonusRestored
		success.BonusRevoked = financials.bonusRevoked
		if err := tx.MarkTicketsRefunded(ctx, ids); err != nil {
			return err
		}
		totalTickets, refundedTickets, err := tx.CountRefundedTickets(ctx, current.Order.ID)
		if err != nil {
			return err
		}
		orderStatus, err := orderStatusAfterRefund(totalTickets, refundedTickets)
		if err != nil {
			return err
		}
		if err := tx.UpdateOrderAfterRefund(ctx, refunds.OrderRefundParams{
			OrderID: current.Order.ID, Status: orderStatus,
			CurrentTotalPrice: financials.currentTotalPrice,
			BonusSpent:        financials.bonusSpent, BonusEarned: financials.bonusEarned,
			PayableAmount: financials.payableAmount,
		}); err != nil {
			return err
		}
		if order.UserID != nil {
			bonusResult, err := tx.ApplyBonus(ctx, refunds.BonusParams{
				UserID: *order.UserID, OrderID: current.Order.ID,
				RefundID: current.ID, RestoreAmount: success.BonusRestored,
				RevokeAmount: success.BonusRevoked,
			})
			if err != nil {
				return err
			}
			success.BonusBalanceAfter = &bonusResult.Balance
			success.BonusDebtAfter = &bonusResult.Debt
		}
		return tx.MarkSuccess(ctx, success)
	})
	if err != nil {
		return nil, mapError(err)
	}
	completed, err := s.repo.GetByKey(ctx, operation.IdempotencyKey)
	if err != nil {
		return nil, mapError(err)
	}
	logger.Info("Возврат завершён: refundID=%d orderID=%d tickets=%d cash=%d",
		completed.ID, completed.Order.ID, len(completed.Items), completed.CashAmount)
	return operationResult(completed), nil
}

func (s *Service) finalizeFailure(
	ctx context.Context,
	operation *refunds.Operation,
	response *supplier.ExecuteRefundResult,
) (*refunds.Result, error) {
	reasonByID := make(map[int]string, len(response.Items))
	for _, item := range response.Items {
		reasonByID[item.TicketID] = item.Reason
	}
	err := s.repo.WithinTransaction(ctx, func(tx refunds.Transaction) error {
		if _, err := tx.LockOrder(ctx, operation.Order.ID); err != nil {
			return err
		}
		current, err := tx.LockOperation(ctx, operation.ID)
		if err != nil {
			return err
		}
		if current.Status != refunds.StatusProcessing {
			return nil
		}
		if err := validateProcessingOperation(current); err != nil {
			return err
		}
		if err := validateSupplierExecution(current, response); err != nil {
			return err
		}
		failure := refunds.FailureParams{
			RefundID: current.ID, SupplierRefundID: response.SupplierRefundID,
			FailureCode: response.FailureCode,
			Items:       make([]refunds.FailureItemParams, 0, len(current.Items)),
		}
		ids := make([]int, 0, len(current.Items))
		for _, item := range current.Items {
			failure.Items = append(failure.Items, refunds.FailureItemParams{
				TicketID: item.ID, Reason: reasonByID[item.ID],
			})
			ids = append(ids, item.ID)
		}
		if err := tx.RestoreTicketsPaid(ctx, ids); err != nil {
			return err
		}
		return tx.MarkFailed(ctx, failure)
	})
	if err != nil {
		return nil, mapError(err)
	}
	failed, err := s.repo.GetByKey(ctx, operation.IdempotencyKey)
	if err != nil {
		return nil, mapError(err)
	}
	logger.Warn("Поставщик отклонил возврат: refundID=%d code=%s", failed.ID, failed.FailureCode)
	return operationResult(failed), nil
}

func orderStatusAfterRefund(total int, refunded int) (string, error) {
	if total <= 0 || refunded <= 0 || refunded > total {
		return "", invalidFinancialState(
			"ticket counts are total=%d refunded=%d",
			total, refunded,
		)
	}
	if total == refunded {
		return orders.OrderStatusRefunded, nil
	}
	return orders.OrderStatusPartiallyRefunded, nil
}
