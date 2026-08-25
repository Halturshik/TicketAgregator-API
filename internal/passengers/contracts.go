package passengers

import (
	"context"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/documents"
)

type Service interface {
	PaymentSynchronizer

	List(ctx context.Context, ownerUserID int) ([]Passenger, error)
	GetOwned(ctx context.Context, ownerUserID int, passengerID int) (*Passenger, error)
	Update(ctx context.Context, ownerUserID int, passengerID int, in SavePassengerInput) (*Passenger, error)
	Delete(ctx context.Context, ownerUserID int, passengerID int) error
}

type Repository interface {
	List(ctx context.Context, ownerUserID int) ([]Passenger, error)
	GetOwned(ctx context.Context, ownerUserID int, passengerID int) (*Passenger, error)
	Update(ctx context.Context, ownerUserID int, passengerID int, in SavePassengerInput, birthDate time.Time) (*Passenger, error)
	Delete(ctx context.Context, ownerUserID int, passengerID int) error
}

type PaymentTransaction interface {
	FindByDocument(ctx context.Context, ownerUserID int, fingerprint string) (int, error)
	InsertFromPayment(ctx context.Context, ownerUserID int, passenger PaymentSnapshot) (int, error)
	UpdateFromPayment(ctx context.Context, ownerUserID int, passengerID int, passenger PaymentSnapshot) error
}

type PaymentSynchronizer interface {
	SyncAfterPayment(
		ctx context.Context,
		passengerTx PaymentTransaction,
		documentTx documents.PaymentTransaction,
		ownerUserID int,
		items []PaymentPassenger,
	) error
}
