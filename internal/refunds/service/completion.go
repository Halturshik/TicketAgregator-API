package service

import (
	"context"

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

		success := refunds.SuccessParams{
			RefundID: operation.ID, SupplierRefundID: response.SupplierRefundID,
			Items: make([]refunds.SuccessItemParams, 0, len(current.Items)),
		}
		ids := make([]int, 0, len(current.Items))
		for _, item := range current.Items {
			supplierItem := responseByID[item.ID]
			cash := supplierItem.RefundAmount - item.BonusSpent
			if cash < 0 || cash > item.PayableAmount {
				return refunds.ErrSupplierMismatch
			}
			success.CashAmount += cash
			success.BonusRestored += item.BonusSpent
			success.BonusRevoked += item.BonusEarned
			success.Items = append(success.Items, refunds.SuccessItemParams{
				TicketID: item.ID, Reason: supplierItem.Reason,
				RefundPercent: supplierItem.RefundPercent, CashRefunded: cash,
				BonusRestored: item.BonusSpent, BonusRevoked: item.BonusEarned,
			})
			ids = append(ids, item.ID)
		}
		if err := tx.MarkTicketsRefunded(ctx, ids); err != nil {
			return err
		}
		if _, err := tx.UpdateOrderStatus(ctx, current.Order.ID); err != nil {
			return err
		}
		if current.Order.UserID != nil {
			bonusResult, err := tx.ApplyBonus(ctx, refunds.BonusParams{
				UserID: *current.Order.UserID, OrderID: current.Order.ID,
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
