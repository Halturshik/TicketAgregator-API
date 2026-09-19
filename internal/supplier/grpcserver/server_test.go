package grpcserver

import (
	"context"
	"testing"

	"github.com/Halturshik/TicketAgregator-API/internal/fare"
	supplierv1 "github.com/Halturshik/TicketAgregator-API/internal/gen/supplier/v1"
	"github.com/Halturshik/TicketAgregator-API/internal/supplier"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestQuoteRefundMapsServiceResult(t *testing.T) {
	policy, _ := fare.Policy(fare.Flexible)
	server := New(&stubSupplierService{quote: &supplier.RefundQuote{
		Items: []supplier.RefundQuoteItemResult{{
			TicketID: 7, Eligible: true, Reason: supplier.RefundReasonAllowed,
			RefundPercent: 70, GrossRefundAmount: 700, Policy: policy,
		}},
	}})
	response, err := server.QuoteRefund(context.Background(), &supplierv1.QuoteRefundRequest{
		ProviderCode: supplier.ProviderAtlas,
		Items:        []*supplierv1.RefundQuoteItem{{TicketId: 7, GrossAmount: 1000}},
	})
	if err != nil {
		t.Fatalf("QuoteRefund() error = %v", err)
	}
	if len(response.Items) != 1 || response.Items[0].TicketId != 7 || response.Items[0].GrossRefundAmount != 700 {
		t.Fatalf("QuoteRefund() response = %+v", response)
	}
}

func TestExecuteRefundMapsIdempotencyConflict(t *testing.T) {
	server := New(&stubSupplierService{executeErr: supplier.ErrIdempotencyConflict})
	_, err := server.ExecuteRefund(context.Background(), &supplierv1.ExecuteRefundRequest{})
	if status.Code(err) != codes.AlreadyExists {
		t.Fatalf("ExecuteRefund() code = %s, want AlreadyExists", status.Code(err))
	}
}

type stubSupplierService struct {
	quote      *supplier.RefundQuote
	executeErr error
}

func (s *stubSupplierService) SearchOffers(context.Context, supplier.SearchRequest) ([]supplier.TripOption, error) {
	return nil, nil
}

func (s *stubSupplierService) QuoteRefund(context.Context, supplier.RefundQuoteRequest) (*supplier.RefundQuote, error) {
	return s.quote, nil
}

func (s *stubSupplierService) ExecuteRefund(context.Context, supplier.ExecuteRefundRequest) (*supplier.ExecuteRefundResult, error) {
	return nil, s.executeErr
}
