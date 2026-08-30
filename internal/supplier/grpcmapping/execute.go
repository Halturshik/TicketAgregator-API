package grpcmapping

import (
	supplierv1 "github.com/Halturshik/TicketAgregator-API/internal/gen/supplier/v1"
	"github.com/Halturshik/TicketAgregator-API/internal/supplier"
)

func ExecuteRefundRequestToProto(request supplier.ExecuteRefundRequest) *supplierv1.ExecuteRefundRequest {
	items := make([]*supplierv1.ExecuteRefundItem, 0, len(request.Items))
	for _, item := range request.Items {
		items = append(items, &supplierv1.ExecuteRefundItem{
			TicketId: int32(item.TicketID), TicketNumber: item.TicketNumber,
			SupplierOfferId: item.SupplierOfferID, FareType: item.FareType,
			DepartureUnix: item.DepartureUnix, GrossAmount: int32(item.GrossAmount),
		})
	}
	return &supplierv1.ExecuteRefundRequest{
		ProviderCode: request.ProviderCode, IdempotencyKey: request.IdempotencyKey, Items: items,
	}
}

func ExecuteRefundRequestFromProto(request *supplierv1.ExecuteRefundRequest) supplier.ExecuteRefundRequest {
	if request == nil {
		return supplier.ExecuteRefundRequest{}
	}
	items := make([]supplier.ExecuteRefundItem, 0, len(request.Items))
	for _, item := range request.Items {
		if item == nil {
			items = append(items, supplier.ExecuteRefundItem{})
			continue
		}
		items = append(items, supplier.ExecuteRefundItem{
			TicketID: int(item.TicketId), TicketNumber: item.TicketNumber,
			SupplierOfferID: item.SupplierOfferId, FareType: item.FareType,
			DepartureUnix: item.DepartureUnix, GrossAmount: int(item.GrossAmount),
		})
	}
	return supplier.ExecuteRefundRequest{
		ProviderCode: request.ProviderCode, IdempotencyKey: request.IdempotencyKey, Items: items,
	}
}

func ExecuteRefundResultToProto(result *supplier.ExecuteRefundResult) *supplierv1.ExecuteRefundResponse {
	items := make([]*supplierv1.ExecuteRefundItemResult, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, &supplierv1.ExecuteRefundItemResult{
			TicketId: int32(item.TicketID), Refunded: item.Refunded, Reason: item.Reason,
			RefundPercent: int32(item.RefundPercent), RefundAmount: int32(item.RefundAmount),
			Policy: PolicyToProto(item.Policy),
		})
	}
	return &supplierv1.ExecuteRefundResponse{
		SupplierRefundId: result.SupplierRefundID, Status: result.Status,
		FailureCode: result.FailureCode, Items: items,
	}
}

func ExecuteRefundResultFromProto(response *supplierv1.ExecuteRefundResponse) *supplier.ExecuteRefundResult {
	items := make([]supplier.ExecuteRefundItemResult, 0, len(response.Items))
	for _, item := range response.Items {
		items = append(items, supplier.ExecuteRefundItemResult{
			TicketID: int(item.TicketId), Refunded: item.Refunded, Reason: item.Reason,
			RefundPercent: int(item.RefundPercent), RefundAmount: int(item.RefundAmount),
			Policy: PolicyFromProto(item.Policy),
		})
	}
	return &supplier.ExecuteRefundResult{
		SupplierRefundID: response.SupplierRefundId, Status: response.Status,
		FailureCode: response.FailureCode, Items: items,
	}
}
