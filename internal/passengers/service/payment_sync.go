package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Halturshik/TicketAgregator-API/internal/documents"
	"github.com/Halturshik/TicketAgregator-API/internal/passengers"
)

func (s *Service) SyncAfterPayment(
	ctx context.Context,
	passengerTx passengers.PaymentTransaction,
	documentTx documents.PaymentTransaction,
	ownerUserID int,
	items []passengers.PaymentPassenger,
) error {
	for _, item := range items {
		if err := syncPaymentPassenger(ctx, passengerTx, documentTx, ownerUserID, item); err != nil {
			return err
		}
	}
	return nil
}

func syncPaymentPassenger(
	ctx context.Context,
	passengerTx passengers.PaymentTransaction,
	documentTx documents.PaymentTransaction,
	ownerUserID int,
	item passengers.PaymentPassenger,
) error {
	switch item.Source {
	case passengers.SourceSelf:
		if item.Document.SavedDocumentID == nil {
			return documentTx.UpsertUserAfterPayment(ctx, ownerUserID, item.Document)
		}
		return nil
	case passengers.SourceNew:
		return saveNewPaymentPassenger(ctx, passengerTx, documentTx, ownerUserID, item)
	case passengers.SourceSaved:
		return updateSavedPaymentPassenger(ctx, passengerTx, documentTx, ownerUserID, item)
	default:
		return fmt.Errorf("unsupported passenger source %q", item.Source)
	}
}

func saveNewPaymentPassenger(
	ctx context.Context,
	passengerTx passengers.PaymentTransaction,
	documentTx documents.PaymentTransaction,
	ownerUserID int,
	item passengers.PaymentPassenger,
) error {
	passengerID, err := passengerTx.FindByDocument(ctx, ownerUserID, item.Document.Fingerprint)
	if err != nil && !errors.Is(err, passengers.ErrNotFound) {
		return err
	}
	if errors.Is(err, passengers.ErrNotFound) {
		passengerID, err = passengerTx.InsertFromPayment(ctx, ownerUserID, item.Passenger)
	} else {
		err = passengerTx.UpdateFromPayment(ctx, ownerUserID, passengerID, item.Passenger)
	}
	if err != nil {
		return err
	}
	return documentTx.UpsertPassengerAfterPayment(ctx, passengerID, item.Document)
}

func updateSavedPaymentPassenger(
	ctx context.Context,
	passengerTx passengers.PaymentTransaction,
	documentTx documents.PaymentTransaction,
	ownerUserID int,
	item passengers.PaymentPassenger,
) error {
	if !item.SaveChanges {
		return nil
	}
	if item.SavedPassengerID == nil {
		return fmt.Errorf("saved passenger source has no passenger id")
	}
	passengerID := *item.SavedPassengerID
	if err := passengerTx.UpdateFromPayment(ctx, ownerUserID, passengerID, item.Passenger); err != nil {
		if errors.Is(err, passengers.ErrNotFound) {
			// Удаление пассажира между созданием и оплатой не должно ломать оплату.
			return nil
		}
		return err
	}
	if item.Document.SavedDocumentID != nil {
		updated, err := documentTx.UpdatePassengerAfterPayment(ctx, passengerID, item.Document)
		if err != nil {
			return err
		}
		if updated {
			return nil
		}
	}
	return documentTx.UpsertPassengerAfterPayment(ctx, passengerID, item.Document)
}
