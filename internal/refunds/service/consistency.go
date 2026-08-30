package service

import (
	"reflect"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/fare"
	"github.com/Halturshik/TicketAgregator-API/internal/refunds"
)

func quoteStillValid(quote *refunds.QuoteOutput, tickets map[int]refunds.TicketData, now time.Time) bool {
	for _, item := range quote.Items {
		ticket, ok := tickets[item.TicketID]
		if !ok {
			return false
		}
		percent, eligible := fare.RefundPercent(ticket.RefundPolicy, ticket.DepartureAt, now)
		if !eligible || percent != item.RefundPercent {
			return false
		}
	}
	return true
}

func sameTickets(expected map[int]refunds.TicketData, actual []refunds.TicketData) bool {
	if len(expected) != len(actual) {
		return false
	}
	for _, ticket := range actual {
		stored, ok := expected[ticket.ID]
		if !ok || !reflect.DeepEqual(stored, ticket) {
			return false
		}
	}
	return true
}

func quoteSupplier(quote *refunds.QuoteOutput) (string, error) {
	code := ""
	for _, item := range quote.Items {
		if code == "" {
			code = item.SupplierCode
		}
		if item.SupplierCode == "" || item.SupplierCode != code {
			return "", refunds.ErrSupplierMismatch
		}
	}
	if code == "" {
		return "", refunds.ErrSupplierMismatch
	}
	return code, nil
}
