package supplier

import "errors"

var (
	ErrInvalidSearchRequest  = errors.New("invalid supplier search request")
	ErrInvalidRefundRequest  = errors.New("invalid supplier refund request")
	ErrSupplierNotFound      = errors.New("supplier not found")
	ErrNoCarriers            = errors.New("supplier carriers not configured")
	ErrIdempotencyConflict   = errors.New("supplier idempotency conflict")
	ErrTicketAlreadyRefunded = errors.New("supplier ticket already refunded")
)
