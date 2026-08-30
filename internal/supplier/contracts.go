package supplier

import "context"

type Gateway interface {
	SearchOffers(ctx context.Context, request SearchRequest) ([]TripOption, error)
	QuoteRefund(ctx context.Context, request RefundQuoteRequest) (*RefundQuote, error)
	ExecuteRefund(ctx context.Context, request ExecuteRefundRequest) (*ExecuteRefundResult, error)
}

type Service interface {
	Gateway
}

type RefundRepository interface {
	CreateOrGet(ctx context.Context, operation RefundOperation) (*RefundOperation, bool, error)
}
