package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
	"strings"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/fare"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
	"github.com/Halturshik/TicketAgregator-API/internal/search/provider"
	"github.com/Halturshik/TicketAgregator-API/internal/supplier"
	"github.com/google/uuid"
)

const maxRefundTickets = 24

func (s *Service) ExecuteRefund(ctx context.Context, request supplier.ExecuteRefundRequest) (*supplier.ExecuteRefundResult, error) {
	requestHash, err := validateExecuteRequest(request)
	if err != nil {
		return nil, err
	}
	operation := supplier.RefundOperation{
		ID: s.newID(), ProviderCode: request.ProviderCode,
		IdempotencyKey: request.IdempotencyKey, RequestHash: requestHash,
		Status: supplier.RefundStatusSuccess,
		Items:  make([]supplier.RefundOperationItem, 0, len(request.Items)),
	}
	for _, item := range request.Items {
		policy, _ := fare.Policy(item.FareType)
		percent, eligible := fare.RefundPercent(policy, time.Unix(item.DepartureUnix, 0).UTC(), s.now().UTC())
		reason := refundReason(policy, eligible)
		operation.Items = append(operation.Items, supplier.RefundOperationItem{
			TicketID: item.TicketID, TicketNumber: item.TicketNumber,
			SupplierOfferID: item.SupplierOfferID, FareType: item.FareType,
			DepartureUnix: item.DepartureUnix, GrossAmount: item.GrossAmount,
			Refunded: eligible, Reason: reason, RefundPercent: percent,
			RefundAmount: item.GrossAmount * percent / 100, Policy: policy,
		})
		if !eligible && operation.Status == supplier.RefundStatusSuccess {
			operation.Status = supplier.RefundStatusRejected
			operation.FailureCode = reason
		}
	}
	if operation.Status == supplier.RefundStatusRejected {
		for index := range operation.Items {
			operation.Items[index].Refunded = false
			operation.Items[index].RefundPercent = 0
			operation.Items[index].RefundAmount = 0
			if operation.Items[index].Reason == supplier.RefundReasonAllowed {
				operation.Items[index].Reason = supplier.RefundReasonBatchRejected
			}
		}
	}

	stored, created, err := s.repo.CreateOrGet(ctx, operation)
	if err != nil {
		return nil, err
	}
	if stored.RequestHash != requestHash {
		return nil, supplier.ErrIdempotencyConflict
	}
	if created && stored.Status == supplier.RefundStatusRejected {
		logger.Warn(
			"Поставщик отклонил возврат: supplier=%s supplierRefundID=%s tickets=%d code=%s",
			stored.ProviderCode, stored.ID, len(stored.Items), stored.FailureCode,
		)
	} else if created {
		logger.Info(
			"Поставщик выполнил возврат: supplier=%s supplierRefundID=%s tickets=%d",
			stored.ProviderCode, stored.ID, len(stored.Items),
		)
	}
	return operationResult(*stored), nil
}

func validateExecuteRequest(request supplier.ExecuteRefundRequest) (string, error) {
	if request.ProviderCode != strings.TrimSpace(request.ProviderCode) ||
		request.IdempotencyKey != strings.TrimSpace(request.IdempotencyKey) {
		return "", supplier.ErrInvalidRefundRequest
	}
	profile, err := provider.Profile(request.ProviderCode)
	if err != nil {
		return "", supplier.ErrSupplierNotFound
	}
	if uuid.Validate(request.IdempotencyKey) != nil ||
		len(request.Items) == 0 || len(request.Items) > maxRefundTickets {
		return "", supplier.ErrInvalidRefundRequest
	}
	seen := make(map[string]struct{}, len(request.Items))
	for _, item := range request.Items {
		if item.TicketNumber != strings.TrimSpace(item.TicketNumber) ||
			item.SupplierOfferID != strings.TrimSpace(item.SupplierOfferID) ||
			item.FareType != strings.TrimSpace(item.FareType) {
			return "", supplier.ErrInvalidRefundRequest
		}
		key := item.SupplierOfferID + ":" + item.TicketNumber
		_, fareSupported := fare.Policy(item.FareType)
		if item.TicketID <= 0 || !validTicketNumber(item.TicketNumber) ||
			uuid.Validate(item.SupplierOfferID) != nil || !fareSupported ||
			!contains(profile.FareTypes, item.FareType) || item.DepartureUnix <= 0 ||
			item.GrossAmount <= 0 {
			return "", supplier.ErrInvalidRefundRequest
		}
		if _, duplicate := seen[key]; duplicate {
			return "", supplier.ErrInvalidRefundRequest
		}
		seen[key] = struct{}{}
	}
	return executeRequestHash(request), nil
}

func validTicketNumber(value string) bool {
	value = strings.TrimSpace(value)
	if len(value) < 6 || len(value) > 20 {
		return false
	}
	for _, character := range value {
		if character != '-' && (character < '0' || character > '9') && (character < 'A' || character > 'Z') {
			return false
		}
	}
	return true
}

func executeRequestHash(request supplier.ExecuteRefundRequest) string {
	items := append([]supplier.ExecuteRefundItem(nil), request.Items...)
	sort.Slice(items, func(i, j int) bool {
		if items[i].TicketID == items[j].TicketID {
			return items[i].TicketNumber < items[j].TicketNumber
		}
		return items[i].TicketID < items[j].TicketID
	})
	payload, _ := json.Marshal(struct {
		ProviderCode string                       `json:"provider_code"`
		Items        []supplier.ExecuteRefundItem `json:"items"`
	}{ProviderCode: request.ProviderCode, Items: items})
	hash := sha256.Sum256(payload)
	return hex.EncodeToString(hash[:])
}

func operationResult(operation supplier.RefundOperation) *supplier.ExecuteRefundResult {
	items := make([]supplier.ExecuteRefundItemResult, 0, len(operation.Items))
	for _, item := range operation.Items {
		items = append(items, supplier.ExecuteRefundItemResult{
			TicketID: item.TicketID, Refunded: item.Refunded, Reason: item.Reason,
			RefundPercent: item.RefundPercent, RefundAmount: item.RefundAmount,
			Policy: item.Policy,
		})
	}
	return &supplier.ExecuteRefundResult{
		SupplierRefundID: operation.ID, Status: operation.Status,
		FailureCode: operation.FailureCode, Items: items,
	}
}
