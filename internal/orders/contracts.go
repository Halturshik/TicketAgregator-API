package orders

import (
	"context"

	"github.com/Halturshik/TicketAgregator-API/internal/documents"
	"github.com/Halturshik/TicketAgregator-API/internal/passengers"
	"github.com/Halturshik/TicketAgregator-API/internal/search"
)

type Service interface {
	Create(ctx context.Context, userID *int, in CreateOrderInput) (*Order, error)
	List(ctx context.Context, userID int, filter ListFilter) ([]Order, error)
}

type Repository interface {
	Create(ctx context.Context, params CreateOrderParams) (*Order, error)
	List(ctx context.Context, userID int, filter ListFilter) ([]Order, error)
}

type SearchReader interface {
	GetCachedResult(ctx context.Context, searchID string) (*search.CachedResult, error)
}

type DocumentValidator interface {
	ValidateForBooking(ctx context.Context, in documents.BookingValidationInput) (*documents.ValidatedBookingDocument, error)
}

type PassengerReader interface {
	GetOwned(ctx context.Context, ownerUserID int, passengerID int) (*passengers.Passenger, error)
}

type PaymentTransaction interface {
	LockForPayment(ctx context.Context, orderID int) (PaymentData, error)
	MarkExpired(ctx context.Context, orderID int) error
	MarkPaid(ctx context.Context, orderID int) error
	ListPassengersForPayment(ctx context.Context, orderID int) ([]passengers.PaymentPassenger, error)
}
