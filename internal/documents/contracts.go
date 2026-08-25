package documents

import (
	"context"
	"time"
)

type Service interface {
	List(ctx context.Context, ownerUserID int, passengerID *int) ([]Document, error)
	Create(ctx context.Context, ownerUserID int, in SaveDocumentInput) (*Document, error)
	Update(ctx context.Context, ownerUserID int, documentID int, in SaveDocumentInput) (*Document, error)
	Delete(ctx context.Context, ownerUserID int, documentID int) error
	ValidateForBooking(ctx context.Context, in BookingValidationInput) (*ValidatedBookingDocument, error)
}

type Repository interface {
	FindRule(ctx context.Context, transport string, isInternational bool, age int, isRussian bool) (*DocumentRule, error)
	ListForUser(ctx context.Context, ownerUserID int) ([]Document, error)
	ListForPassenger(ctx context.Context, ownerUserID int, passengerID int) ([]Document, error)
	CreateForUser(ctx context.Context, ownerUserID int, in SaveDocumentInput, status string, fingerprint string, expiresAt *time.Time, checkedAt time.Time) (*Document, error)
	CreateForPassenger(ctx context.Context, ownerUserID int, passengerID int, in SaveDocumentInput, status string, fingerprint string, expiresAt *time.Time, checkedAt time.Time) (*Document, error)
	GetOwned(ctx context.Context, ownerUserID int, documentID int, passengerID *int) (*Document, error)
	Update(ctx context.Context, ownerUserID int, documentID int, in SaveDocumentInput, status string, fingerprint string, expiresAt *time.Time, checkedAt time.Time) (*Document, error)
	UpdateVerification(ctx context.Context, documentID int, status string, checkedAt time.Time) error
	Delete(ctx context.Context, ownerUserID int, documentID int) error
}

type PaymentTransaction interface {
	UpsertUserAfterPayment(ctx context.Context, ownerUserID int, document PaymentDocument) error
	UpdatePassengerAfterPayment(ctx context.Context, passengerID int, document PaymentDocument) (bool, error)
	UpsertPassengerAfterPayment(ctx context.Context, passengerID int, document PaymentDocument) error
}
