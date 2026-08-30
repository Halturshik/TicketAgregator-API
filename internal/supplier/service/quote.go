package service

import (
	"context"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/fare"
	"github.com/Halturshik/TicketAgregator-API/internal/search/provider"
	"github.com/Halturshik/TicketAgregator-API/internal/supplier"
)

func (s *Service) QuoteRefund(_ context.Context, request supplier.RefundQuoteRequest) (*supplier.RefundQuote, error) {
	profile, err := provider.Profile(request.ProviderCode)
	if err != nil {
		return nil, supplier.ErrSupplierNotFound
	}
	result := make([]supplier.RefundQuoteItemResult, 0, len(request.Items))
	for _, item := range request.Items {
		policy, ok := fare.Policy(item.FareType)
		if !ok || !contains(profile.FareTypes, item.FareType) {
			result = append(result, quoteResult(item, fare.RefundPolicy{}, false, supplier.RefundReasonFareNotSupported, 0))
			continue
		}
		if item.GrossAmount <= 0 || item.DepartureUnix <= 0 || item.SupplierOfferID == "" {
			result = append(result, quoteResult(item, policy, false, supplier.RefundReasonInvalidRequest, 0))
			continue
		}
		percent, eligible := fare.RefundPercent(policy, time.Unix(item.DepartureUnix, 0).UTC(), s.now().UTC())
		reason := refundReason(policy, eligible)
		result = append(result, quoteResult(item, policy, eligible, reason, percent))
	}
	return &supplier.RefundQuote{Items: result}, nil
}

func quoteResult(
	item supplier.RefundQuoteItem,
	policy fare.RefundPolicy,
	eligible bool,
	reason string,
	percent int,
) supplier.RefundQuoteItemResult {
	return supplier.RefundQuoteItemResult{
		TicketID: item.TicketID, Eligible: eligible, Reason: reason,
		RefundPercent: percent, GrossRefundAmount: item.GrossAmount * percent / 100,
		Policy: policy,
	}
}

func refundReason(policy fare.RefundPolicy, eligible bool) string {
	if eligible {
		return supplier.RefundReasonAllowed
	}
	if !policy.Refundable {
		return supplier.RefundReasonNonRefundable
	}
	return supplier.RefundReasonDeadlinePassed
}

func contains(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}
