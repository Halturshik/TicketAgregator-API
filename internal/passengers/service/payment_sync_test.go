package service

import (
	"context"
	"reflect"
	"testing"

	"github.com/Halturshik/TicketAgregator-API/internal/documents"
	"github.com/Halturshik/TicketAgregator-API/internal/passengers"
)

func TestSyncAfterPaymentAppliesPassengerSourcePolicies(t *testing.T) {
	passengerID := 11
	documentID := 22
	tests := []struct {
		name    string
		item    passengers.PaymentPassenger
		prepare func(*fakePaymentPassengerTransaction, *fakePaymentDocumentTransaction)
		want    []string
	}{
		{
			name: "self saves new document",
			item: passengers.PaymentPassenger{Source: passengers.SourceSelf},
			want: []string{"upsert_user_document"},
		},
		{
			name: "self with saved document is untouched",
			item: passengers.PaymentPassenger{
				Source:   passengers.SourceSelf,
				Document: documents.PaymentDocument{SavedDocumentID: &documentID},
			},
			want: nil,
		},
		{
			name: "new passenger is inserted and gets document",
			item: passengers.PaymentPassenger{Source: passengers.SourceNew},
			prepare: func(passengerTx *fakePaymentPassengerTransaction, _ *fakePaymentDocumentTransaction) {
				passengerTx.findErr = passengers.ErrNotFound
				passengerTx.insertedPassengerID = passengerID
			},
			want: []string{"find_passenger", "insert_passenger", "upsert_passenger_document"},
		},
		{
			name: "new passenger matching a document is updated",
			item: passengers.PaymentPassenger{Source: passengers.SourceNew},
			prepare: func(passengerTx *fakePaymentPassengerTransaction, _ *fakePaymentDocumentTransaction) {
				passengerTx.findPassengerID = passengerID
			},
			want: []string{"find_passenger", "update_passenger", "upsert_passenger_document"},
		},
		{
			name: "saved passenger without changes is untouched",
			item: passengers.PaymentPassenger{Source: passengers.SourceSaved, SavedPassengerID: &passengerID},
			want: nil,
		},
		{
			name: "saved passenger updates existing document",
			item: passengers.PaymentPassenger{
				Source: passengers.SourceSaved, SavedPassengerID: &passengerID, SaveChanges: true,
				Document: documents.PaymentDocument{SavedDocumentID: &documentID},
			},
			prepare: func(_ *fakePaymentPassengerTransaction, documentTx *fakePaymentDocumentTransaction) {
				documentTx.updated = true
			},
			want: []string{"update_passenger", "update_passenger_document"},
		},
		{
			name: "saved passenger recreates missing document",
			item: passengers.PaymentPassenger{
				Source: passengers.SourceSaved, SavedPassengerID: &passengerID, SaveChanges: true,
				Document: documents.PaymentDocument{SavedDocumentID: &documentID},
			},
			want: []string{"update_passenger", "update_passenger_document", "upsert_passenger_document"},
		},
		{
			name: "deleted saved passenger is ignored",
			item: passengers.PaymentPassenger{
				Source: passengers.SourceSaved, SavedPassengerID: &passengerID, SaveChanges: true,
			},
			prepare: func(passengerTx *fakePaymentPassengerTransaction, _ *fakePaymentDocumentTransaction) {
				passengerTx.updateErr = passengers.ErrNotFound
			},
			want: []string{"update_passenger"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var calls []string
			passengerTx := &fakePaymentPassengerTransaction{calls: &calls}
			documentTx := &fakePaymentDocumentTransaction{calls: &calls}
			if test.prepare != nil {
				test.prepare(passengerTx, documentTx)
			}
			svc := &Service{}
			if err := svc.SyncAfterPayment(
				context.Background(), passengerTx, documentTx, 7, []passengers.PaymentPassenger{test.item},
			); err != nil {
				t.Fatalf("SyncAfterPayment() error = %v", err)
			}
			if !reflect.DeepEqual(calls, test.want) {
				t.Fatalf("calls = %v, want %v", calls, test.want)
			}
		})
	}
}

type fakePaymentPassengerTransaction struct {
	calls               *[]string
	findPassengerID     int
	findErr             error
	insertedPassengerID int
	insertErr           error
	updateErr           error
}

func (f *fakePaymentPassengerTransaction) call(name string) { *f.calls = append(*f.calls, name) }

func (f *fakePaymentPassengerTransaction) FindByDocument(context.Context, int, string) (int, error) {
	f.call("find_passenger")
	return f.findPassengerID, f.findErr
}

func (f *fakePaymentPassengerTransaction) InsertFromPayment(context.Context, int, passengers.PaymentSnapshot) (int, error) {
	f.call("insert_passenger")
	return f.insertedPassengerID, f.insertErr
}

func (f *fakePaymentPassengerTransaction) UpdateFromPayment(context.Context, int, int, passengers.PaymentSnapshot) error {
	f.call("update_passenger")
	return f.updateErr
}

type fakePaymentDocumentTransaction struct {
	calls     *[]string
	updated   bool
	updateErr error
	upsertErr error
}

func (f *fakePaymentDocumentTransaction) call(name string) { *f.calls = append(*f.calls, name) }

func (f *fakePaymentDocumentTransaction) UpsertUserAfterPayment(context.Context, int, documents.PaymentDocument) error {
	f.call("upsert_user_document")
	return f.upsertErr
}

func (f *fakePaymentDocumentTransaction) UpdatePassengerAfterPayment(context.Context, int, documents.PaymentDocument) (bool, error) {
	f.call("update_passenger_document")
	return f.updated, f.updateErr
}

func (f *fakePaymentDocumentTransaction) UpsertPassengerAfterPayment(context.Context, int, documents.PaymentDocument) error {
	f.call("upsert_passenger_document")
	return f.upsertErr
}
