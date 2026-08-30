package refunds

import "errors"

var (
	ErrNotFound            = errors.New("refund resource not found")
	ErrForbidden           = errors.New("refund forbidden")
	ErrInvalidStatus       = errors.New("invalid refund status")
	ErrTicketsMismatch     = errors.New("refund tickets mismatch")
	ErrNotAllowed          = errors.New("refund not allowed")
	ErrSupplierMismatch    = errors.New("supplier refund quote mismatch")
	ErrIdempotencyConflict = errors.New("refund idempotency conflict")
)
