package service

import (
	"context"
	"errors"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
	"github.com/Halturshik/TicketAgregator-API/internal/refunds"
)

func (s *Service) createOperation(
	ctx context.Context,
	userID *int,
	orderID int,
	input refunds.Input,
	requestHash string,
	supplierCode string,
	reconciliationStartedAt time.Time,
	prepared *preparedRefund,
) (bool, error) {
	created := false
	err := s.repo.WithinTransaction(ctx, func(tx refunds.Transaction) error {
		order, err := tx.LockOrder(ctx, orderID)
		if err != nil {
			return err
		}
		if err := authorize(order, userID, input.GuestPaymentToken); err != nil {
			return err
		}
		if _, err := tx.GetByKey(ctx, input.IdempotencyKey); err == nil {
			return nil
		} else if !errors.Is(err, refunds.ErrNotFound) {
			return err
		}
		if err := ensureRefundableOrder(order); err != nil {
			return err
		}
		if !sameOrderFinancials(prepared.order, order) {
			return refunds.ErrTicketsMismatch
		}
		lockedTickets, err := tx.LockTickets(ctx, orderID, input.TicketIDs, input.All)
		if err != nil {
			return err
		}
		if err := ensurePaidTickets(lockedTickets, len(input.TicketIDs), input.All); err != nil {
			return err
		}
		if !sameTickets(prepared.tickets, lockedTickets) {
			return refunds.ErrTicketsMismatch
		}
		if !quoteStillValid(prepared.quote, prepared.tickets, s.now()) {
			return refunds.ErrNotAllowed
		}
		paymentID, err := tx.SuccessfulPaymentID(ctx, orderID)
		if err != nil {
			return err
		}
		refundID, wasCreated, err := tx.CreateProcessing(ctx, createParams(
			orderID, paymentID, input.IdempotencyKey, requestHash,
			supplierCode, reconciliationStartedAt,
			prepared.quote, prepared.tickets,
		))
		if err != nil {
			return err
		}
		if !wasCreated {
			return nil
		}
		created = true
		if err := tx.MarkTicketsPending(ctx, ticketIDs(lockedTickets)); err != nil {
			return err
		}
		logger.Info("Возврат подготовлен: refundID=%d orderID=%d tickets=%d", refundID, orderID, len(lockedTickets))
		return nil
	})
	return created, err
}

func createParams(
	orderID int,
	paymentID int,
	idempotencyKey string,
	requestHash string,
	supplierCode string,
	reconciliationStartedAt time.Time,
	quote *refunds.QuoteOutput,
	tickets map[int]refunds.TicketData,
) refunds.CreateParams {
	items := make([]refunds.CreateItemParams, 0, len(quote.Items))
	for _, item := range quote.Items {
		ticket := tickets[item.TicketID]
		items = append(items, refunds.CreateItemParams{
			TicketID: ticket.ID, SupplierCode: ticket.SupplierCode,
			TicketNumber: ticket.TicketNumber, SupplierOfferID: ticket.SupplierOfferID,
			FareType: ticket.FareType, DepartureAt: ticket.DepartureAt,
			Reason: item.Reason, RefundPercent: item.RefundPercent,
			GrossAmount: item.GrossAmount, SupplierRefundAmount: item.GrossRefundAmount,
		})
	}
	return refunds.CreateParams{
		OrderID: orderID, PaymentID: paymentID, IdempotencyKey: idempotencyKey,
		RequestHash: requestHash, SupplierCode: supplierCode,
		NextRetryAt:            reconciliationStartedAt,
		ReconciliationDeadline: reconciliationStartedAt.Add(refunds.ReconciliationWindow),
		CashAmount:             quote.CashAmount, BonusRestored: quote.BonusRestored,
		BonusRevoked: quote.BonusRevoked, Items: items,
	}
}

func sameOrderFinancials(expected *refunds.OrderData, actual *refunds.OrderData) bool {
	return expected != nil && actual != nil &&
		expected.Status == actual.Status &&
		expected.CurrentTotalPrice == actual.CurrentTotalPrice &&
		expected.BonusSpent == actual.BonusSpent &&
		expected.BonusEarned == actual.BonusEarned &&
		expected.PayableAmount == actual.PayableAmount
}

func ticketIDs(tickets []refunds.TicketData) []int {
	ids := make([]int, len(tickets))
	for index, ticket := range tickets {
		ids[index] = ticket.ID
	}
	return ids
}
