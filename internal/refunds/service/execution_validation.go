package service

import (
	"reflect"

	"github.com/Halturshik/TicketAgregator-API/internal/fare"
	"github.com/Halturshik/TicketAgregator-API/internal/orders"
	"github.com/Halturshik/TicketAgregator-API/internal/refunds"
	"github.com/Halturshik/TicketAgregator-API/internal/supplier"
	"github.com/google/uuid"
)

func validateProcessingOperation(operation *refunds.Operation) error {
	if operation == nil || operation.Status != refunds.StatusProcessing || len(operation.Items) == 0 {
		return refunds.ErrSupplierMismatch
	}
	for _, item := range operation.Items {
		if item.Status != orders.TicketStatusRefundPending || item.SupplierCode != operation.SupplierCode {
			return refunds.ErrSupplierMismatch
		}
	}
	return nil
}

func validateSupplierExecution(operation *refunds.Operation, response *supplier.ExecuteRefundResult) error {
	if response == nil || uuid.Validate(response.SupplierRefundID) != nil ||
		len(response.Items) != len(operation.Items) {
		return refunds.ErrSupplierMismatch
	}
	if response.Status != supplier.RefundStatusSuccess && response.Status != supplier.RefundStatusRejected {
		return refunds.ErrSupplierMismatch
	}
	if response.Status == supplier.RefundStatusSuccess && response.FailureCode != "" ||
		response.Status == supplier.RefundStatusRejected && response.FailureCode == "" {
		return refunds.ErrSupplierMismatch
	}

	byID := make(map[int]refunds.OperationItem, len(operation.Items))
	for _, item := range operation.Items {
		byID[item.ID] = item
	}
	seen := make(map[int]struct{}, len(response.Items))
	for _, result := range response.Items {
		item, exists := byID[result.TicketID]
		if !exists {
			return refunds.ErrSupplierMismatch
		}
		if _, duplicate := seen[result.TicketID]; duplicate {
			return refunds.ErrSupplierMismatch
		}
		seen[result.TicketID] = struct{}{}
		if !reflect.DeepEqual(result.Policy, item.RefundPolicy) {
			return refunds.ErrSupplierMismatch
		}
		if response.Status == supplier.RefundStatusSuccess {
			if !result.Refunded || result.Reason != supplier.RefundReasonAllowed ||
				!allowedRefundPercent(item.RefundPolicy, result.RefundPercent) ||
				result.RefundAmount != item.GrossAmount*result.RefundPercent/100 {
				return refunds.ErrSupplierMismatch
			}
			continue
		}
		if result.Refunded || result.RefundPercent != 0 || result.RefundAmount != 0 ||
			!allowedFailureReason(result.Reason) {
			return refunds.ErrSupplierMismatch
		}
	}
	return nil
}

func allowedRefundPercent(policy fare.RefundPolicy, percent int) bool {
	if !policy.Refundable || percent <= 0 || percent > 100 {
		return false
	}
	for _, tier := range policy.Tiers {
		if tier.RefundPercent == percent {
			return true
		}
	}
	return false
}

func allowedFailureReason(reason string) bool {
	switch reason {
	case supplier.RefundReasonNonRefundable, supplier.RefundReasonDeadlinePassed,
		supplier.RefundReasonFareNotSupported, supplier.RefundReasonInvalidRequest,
		supplier.RefundReasonBatchRejected:
		return true
	default:
		return false
	}
}
