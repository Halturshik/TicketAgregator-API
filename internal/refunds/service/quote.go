package service

import (
	"context"
	"reflect"
	"sort"

	"github.com/Halturshik/TicketAgregator-API/internal/fare"
	"github.com/Halturshik/TicketAgregator-API/internal/refunds"
	"github.com/Halturshik/TicketAgregator-API/internal/supplier"
)

type preparedRefund struct {
	order   *refunds.OrderData
	tickets map[int]refunds.TicketData
	quote   *refunds.QuoteOutput
}

func (s *Service) Quote(ctx context.Context, userID *int, orderID int, input refunds.Input) (*refunds.QuoteOutput, error) {
	prepared, err := s.prepare(ctx, userID, orderID, input)
	if err != nil {
		return nil, mapError(err)
	}
	return prepared.quote, nil
}

func (s *Service) prepare(ctx context.Context, userID *int, orderID int, input refunds.Input) (*preparedRefund, error) {
	if err := validateSelection(orderID, input); err != nil {
		return nil, err
	}
	order, tickets, err := s.repo.Load(ctx, orderID, input.TicketIDs, input.All)
	if err != nil {
		return nil, err
	}
	if err := authorize(order, userID, input.GuestPaymentToken); err != nil {
		return nil, err
	}
	if err := ensureRefundableOrder(order); err != nil {
		return nil, err
	}
	if err := ensurePaidTickets(tickets, len(input.TicketIDs), input.All); err != nil {
		return nil, err
	}

	byProvider := make(map[string][]supplier.RefundQuoteItem)
	ticketByID := make(map[int]refunds.TicketData, len(tickets))
	items := make([]refunds.QuoteItem, 0, len(tickets))
	seenResults := make(map[int]struct{}, len(tickets))
	for _, ticket := range tickets {
		ticketByID[ticket.ID] = ticket
		if ticket.SupplierCode == supplier.ProviderLegacy {
			if ticket.RefundPolicy.Refundable || ticket.FareType != fare.NonRefundable {
				return nil, refunds.ErrSupplierMismatch
			}
			items = append(items, refunds.QuoteItem{
				TicketID: ticket.ID, SupplierCode: ticket.SupplierCode,
				Refundable: false, Reason: supplier.RefundReasonNonRefundable,
				GrossAmount: ticket.Price,
			})
			seenResults[ticket.ID] = struct{}{}
			continue
		}
		byProvider[ticket.SupplierCode] = append(byProvider[ticket.SupplierCode], supplier.RefundQuoteItem{
			TicketID: ticket.ID, SupplierOfferID: ticket.SupplierOfferID, FareType: ticket.FareType,
			DepartureUnix: ticket.DepartureAt.Unix(), GrossAmount: ticket.Price,
		})
	}

	for providerCode, requestItems := range byProvider {
		quote, err := s.supplier.QuoteRefund(ctx, supplier.RefundQuoteRequest{
			ProviderCode: providerCode,
			Items:        requestItems,
		})
		if err != nil {
			return nil, err
		}
		for _, result := range quote.Items {
			ticket, exists := ticketByID[result.TicketID]
			if !exists || ticket.SupplierCode != providerCode {
				return nil, refunds.ErrSupplierMismatch
			}
			if _, duplicate := seenResults[result.TicketID]; duplicate {
				return nil, refunds.ErrSupplierMismatch
			}
			seenResults[result.TicketID] = struct{}{}
			if !reflect.DeepEqual(result.Policy, ticket.RefundPolicy) {
				return nil, refunds.ErrSupplierMismatch
			}
			if result.RefundPercent < 0 || result.RefundPercent > 100 ||
				result.Eligible != (result.RefundPercent > 0) {
				return nil, refunds.ErrSupplierMismatch
			}
			expectedGross := ticket.Price * result.RefundPercent / 100
			if result.GrossRefundAmount != expectedGross {
				return nil, refunds.ErrSupplierMismatch
			}
			item := refunds.QuoteItem{
				TicketID: result.TicketID, SupplierCode: ticket.SupplierCode,
				Refundable: result.Eligible, Reason: result.Reason,
				RefundPercent: result.RefundPercent, GrossAmount: ticket.Price,
				GrossRefundAmount: result.GrossRefundAmount,
			}
			items = append(items, item)
		}
	}
	if len(items) != len(tickets) || len(seenResults) != len(tickets) {
		return nil, refunds.ErrSupplierMismatch
	}
	sort.Slice(items, func(i, j int) bool { return items[i].TicketID < items[j].TicketID })

	output := &refunds.QuoteOutput{OrderID: orderID, Refundable: true, Items: items}
	refundedTicketTotal := 0
	supplierRefundAmount := 0
	for _, item := range items {
		if !item.Refundable {
			output.Refundable = false
			if output.Reason == "" {
				output.Reason = item.Reason
			}
			continue
		}
		refundedTicketTotal += item.GrossAmount
		supplierRefundAmount += item.GrossRefundAmount
	}
	if output.Refundable {
		financials, err := calculateRefundFinancials(order, refundedTicketTotal, supplierRefundAmount)
		if err != nil {
			return nil, err
		}
		output.CashAmount = financials.cashAmount
		output.BonusRestored = financials.bonusRestored
		output.BonusRevoked = financials.bonusRevoked
	}
	return &preparedRefund{order: order, tickets: ticketByID, quote: output}, nil
}
