package service

import (
	"github.com/Halturshik/TicketAgregator-API/internal/bookingaccess"
	"github.com/Halturshik/TicketAgregator-API/internal/refunds"
)

func refundQuote(
	booking *bookingaccess.Booking,
	quote *refunds.QuoteOutput,
) *bookingaccess.RefundQuote {
	return &bookingaccess.RefundQuote{
		OrderNumber:   booking.OrderNumber,
		Refundable:    quote.Refundable,
		Reason:        quote.Reason,
		CashAmount:    quote.CashAmount,
		BonusRestored: quote.BonusRestored,
		BonusRevoked:  quote.BonusRevoked,
		Items:         refundItems(booking, quote.Items),
	}
}

func refundResult(
	booking *bookingaccess.Booking,
	result *refunds.Result,
) *bookingaccess.RefundResult {
	return &bookingaccess.RefundResult{
		OrderNumber:      booking.OrderNumber,
		OrderStatus:      result.OrderStatus,
		Status:           result.Status,
		SupplierRefundID: result.SupplierRefundID,
		FailureCode:      result.FailureCode,
		AttemptCount:     result.AttemptCount,
		NextRetryAt:      result.NextRetryAt,
		CashAmount:       result.CashAmount,
		BonusRestored:    result.BonusRestored,
		BonusRevoked:     result.BonusRevoked,
		BonusBalance:     result.BonusBalance,
		BonusDebt:        result.BonusDebt,
		Items:            refundItems(booking, result.Items),
	}
}

func refundItems(
	booking *bookingaccess.Booking,
	items []refunds.QuoteItem,
) []bookingaccess.RefundQuoteItem {
	numbers := make(map[int]string, len(booking.Tickets))
	for _, ticket := range booking.Tickets {
		numbers[ticket.ID] = ticket.Number
	}
	result := make([]bookingaccess.RefundQuoteItem, 0, len(items))
	for _, item := range items {
		result = append(result, bookingaccess.RefundQuoteItem{
			TicketNumber:      numbers[item.TicketID],
			Refundable:        item.Refundable,
			Reason:            item.Reason,
			RefundPercent:     item.RefundPercent,
			GrossAmount:       item.GrossAmount,
			GrossRefundAmount: item.GrossRefundAmount,
		})
	}
	return result
}
