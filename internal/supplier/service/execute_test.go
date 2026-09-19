package service

import (
	"context"
	"testing"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/fare"
	"github.com/Halturshik/TicketAgregator-API/internal/supplier"
)

const (
	testIdempotencyKey = "11111111-1111-4111-8111-111111111111"
	testOfferID        = "22222222-2222-4222-8222-222222222222"
	testRefundID       = "33333333-3333-4333-8333-333333333333"
)

func TestExecuteRefundIsIdempotentForSameCommand(t *testing.T) {
	now := time.Date(2026, time.August, 29, 12, 0, 0, 0, time.UTC)
	repo := &memoryRefundRepository{}
	service := &Service{repo: repo, now: func() time.Time { return now }, newID: func() string { return testRefundID }}
	request := validExecuteRequest(now)

	first, err := service.ExecuteRefund(context.Background(), request)
	if err != nil {
		t.Fatalf("first ExecuteRefund() error = %v", err)
	}
	second, err := service.ExecuteRefund(context.Background(), request)
	if err != nil {
		t.Fatalf("second ExecuteRefund() error = %v", err)
	}
	if repo.created != 1 || first.SupplierRefundID != second.SupplierRefundID || first.Items[0].RefundAmount != 900 {
		t.Fatalf("idempotency result: created=%d first=%+v second=%+v", repo.created, first, second)
	}
}

func TestExecuteRefundRejectsSameKeyForDifferentCommand(t *testing.T) {
	now := time.Date(2026, time.August, 29, 12, 0, 0, 0, time.UTC)
	service := &Service{
		repo: &memoryRefundRepository{}, now: func() time.Time { return now },
		newID: func() string { return testRefundID },
	}
	request := validExecuteRequest(now)
	if _, err := service.ExecuteRefund(context.Background(), request); err != nil {
		t.Fatalf("first ExecuteRefund() error = %v", err)
	}
	request.Items[0].GrossAmount++
	if _, err := service.ExecuteRefund(context.Background(), request); err != supplier.ErrIdempotencyConflict {
		t.Fatalf("second ExecuteRefund() error = %v, want idempotency conflict", err)
	}
}

func TestExecuteRefundRejectsWholeBatchAfterDeadline(t *testing.T) {
	now := time.Date(2026, time.August, 29, 12, 0, 0, 0, time.UTC)
	request := validExecuteRequest(now)
	request.Items = append(request.Items, supplier.ExecuteRefundItem{
		TicketID: 2, TicketNumber: "AB-12300002", SupplierOfferID: testOfferID,
		FareType: fare.Flexible, DepartureUnix: now.Add(23 * time.Hour).Unix(), GrossAmount: 1000,
	})
	service := &Service{
		repo: &memoryRefundRepository{}, now: func() time.Time { return now },
		newID: func() string { return testRefundID },
	}
	result, err := service.ExecuteRefund(context.Background(), request)
	if err != nil {
		t.Fatalf("ExecuteRefund() error = %v", err)
	}
	if result.Status != supplier.RefundStatusRejected || result.FailureCode != supplier.RefundReasonDeadlinePassed {
		t.Fatalf("result = %+v", result)
	}
	if result.Items[0].Refunded || result.Items[0].Reason != supplier.RefundReasonBatchRejected || result.Items[1].Refunded {
		t.Fatalf("batch items = %+v", result.Items)
	}
}

func validExecuteRequest(now time.Time) supplier.ExecuteRefundRequest {
	return supplier.ExecuteRefundRequest{
		ProviderCode: supplier.ProviderAtlas, IdempotencyKey: testIdempotencyKey,
		Items: []supplier.ExecuteRefundItem{{
			TicketID: 1, TicketNumber: "AB-12300001", SupplierOfferID: testOfferID,
			FareType: fare.Flexible, DepartureUnix: now.Add(10 * 24 * time.Hour).Unix(), GrossAmount: 1000,
		}},
	}
}

type memoryRefundRepository struct {
	operation *supplier.RefundOperation
	created   int
}

func (r *memoryRefundRepository) CreateOrGet(_ context.Context, operation supplier.RefundOperation) (*supplier.RefundOperation, bool, error) {
	if r.operation != nil {
		copy := *r.operation
		copy.Items = append([]supplier.RefundOperationItem(nil), r.operation.Items...)
		return &copy, false, nil
	}
	r.created++
	copy := operation
	copy.Items = append([]supplier.RefundOperationItem(nil), operation.Items...)
	r.operation = &copy
	return &operation, true, nil
}
