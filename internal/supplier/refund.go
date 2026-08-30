package supplier

import "github.com/Halturshik/TicketAgregator-API/internal/fare"

const (
	RefundStatusSuccess  = "success"
	RefundStatusRejected = "rejected"

	RefundReasonAllowed          = "allowed"
	RefundReasonNonRefundable    = "non_refundable"
	RefundReasonDeadlinePassed   = "deadline_passed"
	RefundReasonFareNotSupported = "fare_not_supported"
	RefundReasonInvalidRequest   = "invalid_request"
	RefundReasonBatchRejected    = "batch_rejected"
)

type RefundQuoteRequest struct {
	ProviderCode string
	Items        []RefundQuoteItem
}

type RefundQuoteItem struct {
	TicketID        int
	SupplierOfferID string
	FareType        string
	DepartureUnix   int64
	GrossAmount     int
}

type RefundQuote struct {
	Items []RefundQuoteItemResult
}

type RefundQuoteItemResult struct {
	TicketID          int
	Eligible          bool
	Reason            string
	RefundPercent     int
	GrossRefundAmount int
	Policy            fare.RefundPolicy
}

type ExecuteRefundRequest struct {
	ProviderCode   string
	IdempotencyKey string
	Items          []ExecuteRefundItem
}

type ExecuteRefundItem struct {
	TicketID        int
	TicketNumber    string
	SupplierOfferID string
	FareType        string
	DepartureUnix   int64
	GrossAmount     int
}

type ExecuteRefundResult struct {
	SupplierRefundID string
	Status           string
	FailureCode      string
	Items            []ExecuteRefundItemResult
}

type ExecuteRefundItemResult struct {
	TicketID      int
	Refunded      bool
	Reason        string
	RefundPercent int
	RefundAmount  int
	Policy        fare.RefundPolicy
}
