package grpcmapping

import (
	supplierv1 "github.com/Halturshik/TicketAgregator-API/internal/gen/supplier/v1"
	"github.com/Halturshik/TicketAgregator-API/internal/supplier"
)

func RefundRequestToProto(request supplier.RefundQuoteRequest) *supplierv1.QuoteRefundRequest {
	items := make([]*supplierv1.RefundQuoteItem, 0, len(request.Items))
	for _, item := range request.Items {
		items = append(items, &supplierv1.RefundQuoteItem{
			TicketId: int32(item.TicketID), SupplierOfferId: item.SupplierOfferID,
			FareType: item.FareType, DepartureUnix: item.DepartureUnix, GrossAmount: int32(item.GrossAmount),
		})
	}
	return &supplierv1.QuoteRefundRequest{ProviderCode: request.ProviderCode, Items: items}
}

func RefundRequestFromProto(request *supplierv1.QuoteRefundRequest) supplier.RefundQuoteRequest {
	if request == nil {
		return supplier.RefundQuoteRequest{}
	}
	items := make([]supplier.RefundQuoteItem, 0, len(request.Items))
	for _, item := range request.Items {
		if item == nil {
			items = append(items, supplier.RefundQuoteItem{})
			continue
		}
		items = append(items, supplier.RefundQuoteItem{
			TicketID: int(item.TicketId), SupplierOfferID: item.SupplierOfferId,
			FareType: item.FareType, DepartureUnix: item.DepartureUnix,
			GrossAmount: int(item.GrossAmount),
		})
	}
	return supplier.RefundQuoteRequest{ProviderCode: request.ProviderCode, Items: items}
}

func RefundQuoteFromProto(response *supplierv1.QuoteRefundResponse) *supplier.RefundQuote {
	items := make([]supplier.RefundQuoteItemResult, 0, len(response.Items))
	for _, item := range response.Items {
		items = append(items, supplier.RefundQuoteItemResult{
			TicketID: int(item.TicketId), Eligible: item.Eligible, Reason: item.Reason,
			RefundPercent: int(item.RefundPercent), GrossRefundAmount: int(item.GrossRefundAmount),
			Policy: PolicyFromProto(item.Policy),
		})
	}
	return &supplier.RefundQuote{Items: items}
}

func RefundQuoteToProto(quote *supplier.RefundQuote) *supplierv1.QuoteRefundResponse {
	items := make([]*supplierv1.RefundQuoteItemResult, 0, len(quote.Items))
	for _, item := range quote.Items {
		items = append(items, &supplierv1.RefundQuoteItemResult{
			TicketId: int32(item.TicketID), Eligible: item.Eligible, Reason: item.Reason,
			RefundPercent: int32(item.RefundPercent), GrossRefundAmount: int32(item.GrossRefundAmount),
			Policy: PolicyToProto(item.Policy),
		})
	}
	return &supplierv1.QuoteRefundResponse{Items: items}
}
