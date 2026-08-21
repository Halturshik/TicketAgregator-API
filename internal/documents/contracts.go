package documents

import "context"

type Service interface {
	List(ctx context.Context, ownerUserID int, passengerID *int) ([]Document, error)
	Create(ctx context.Context, ownerUserID int, in SaveDocumentInput) (*Document, error)
	ValidateForBooking(ctx context.Context, in BookingValidationInput) error
}
