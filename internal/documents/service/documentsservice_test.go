package service

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/documents"
)

func TestVerificationIsStableWithinPeriodAndNormalizesNumber(t *testing.T) {
	secret := []byte("test-secret")
	checkedAt := time.Date(2026, time.August, 10, 12, 0, 0, 0, time.UTC)
	statusA, fingerprintA := verifyDocument(secret, "internal_passport", "1234 567890", checkedAt)
	statusB, fingerprintB := verifyDocument(secret, "internal_passport", "1234567890", checkedAt.Add(20*24*time.Hour))

	if statusA != statusB {
		t.Fatalf("status changed in one period: %s -> %s", statusA, statusB)
	}
	if fingerprintA != fingerprintB {
		t.Fatal("equivalent document numbers have different fingerprints")
	}
}

func TestVerificationDistributionIsCloseToFivePercent(t *testing.T) {
	secret := []byte("distribution-secret")
	checkedAt := time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC)
	const samples = 10000
	rejected := 0
	for i := 0; i < samples; i++ {
		status, _ := verifyDocument(secret, "internal_passport", fmt.Sprintf("%010d", i), checkedAt)
		if status == documents.StatusRejected {
			rejected++
		}
	}
	percentage := float64(rejected) * 100 / samples
	if percentage < 4 || percentage > 6 {
		t.Fatalf("rejected percentage = %.2f, want close to 5", percentage)
	}
}

func TestNonVerifiableDocumentTypesAreAlwaysVerified(t *testing.T) {
	for _, documentType := range []string{"birth_certificate", "foreign_passport"} {
		for month := 1; month <= 12; month++ {
			status, _ := verifyDocument([]byte("secret"), documentType, "ABC12345", time.Date(2026, time.Month(month), 1, 0, 0, 0, 0, time.UTC))
			if status != documents.StatusVerified {
				t.Fatalf("%s status = %s, want verified", documentType, status)
			}
		}
	}
}

func TestValidateForBookingRechecksSavedDocumentAndPersistsRejection(t *testing.T) {
	now := time.Date(2026, time.August, 10, 12, 0, 0, 0, time.UTC)
	secret := []byte("booking-secret")
	number := findDocumentWithStatus(t, secret, now, documents.StatusRejected)
	repo := &fakeDocumentRepository{
		rule: &documents.DocumentRule{AllowInternalPassport: true},
		doc:  &documents.Document{ID: 44, Type: "internal_passport", Number: number},
	}
	svc := &Service{repo: repo, verificationSecret: secret, now: func() time.Time { return now }}
	documentID := 44
	ownerID := 7

	_, err := svc.ValidateForBooking(context.Background(), validBookingInput(&ownerID, nil, &documentID))
	if !errors.Is(err, apierror.ErrDocumentRejected) {
		t.Fatalf("ValidateForBooking() error = %v, want document rejected", err)
	}
	if repo.updatedDocumentID != documentID || repo.updatedStatus != documents.StatusRejected {
		t.Fatalf("persisted verification = id:%d status:%s", repo.updatedDocumentID, repo.updatedStatus)
	}
}

func TestValidateForBookingReturnsVerifiedSnapshot(t *testing.T) {
	now := time.Date(2026, time.August, 10, 12, 0, 0, 0, time.UTC)
	secret := []byte("booking-secret")
	number := findDocumentWithStatus(t, secret, now, documents.StatusVerified)
	repo := &fakeDocumentRepository{rule: &documents.DocumentRule{AllowInternalPassport: true}}
	svc := &Service{repo: repo, verificationSecret: secret, now: func() time.Time { return now }}
	in := validBookingInput(nil, nil, nil)
	in.Document.Type = "internal_passport"
	in.Document.Number = number

	doc, err := svc.ValidateForBooking(context.Background(), in)
	if err != nil {
		t.Fatalf("ValidateForBooking() error = %v", err)
	}
	if doc.VerificationStatus != documents.StatusVerified || doc.Fingerprint == "" || doc.LastCheckedAt == "" {
		t.Fatalf("validated document = %+v", doc)
	}
}

func TestSavedDocumentIsRecheckedInNextPeriod(t *testing.T) {
	secret := []byte("period-transition-secret")
	august := time.Date(2026, time.August, 10, 12, 0, 0, 0, time.UTC)
	september := time.Date(2026, time.September, 10, 12, 0, 0, 0, time.UTC)
	number := findDocumentTransition(t, secret, august, documents.StatusVerified, september, documents.StatusRejected)
	repo := &fakeDocumentRepository{
		rule: &documents.DocumentRule{AllowInternalPassport: true},
		doc:  &documents.Document{ID: 45, Type: "internal_passport", Number: number},
	}
	current := august
	svc := &Service{repo: repo, verificationSecret: secret, now: func() time.Time { return current }}
	documentID := 45
	ownerID := 7

	if _, err := svc.ValidateForBooking(context.Background(), validBookingInput(&ownerID, nil, &documentID)); err != nil {
		t.Fatalf("first verification error = %v", err)
	}
	current = september
	if _, err := svc.ValidateForBooking(context.Background(), validBookingInput(&ownerID, nil, &documentID)); !errors.Is(err, apierror.ErrDocumentRejected) {
		t.Fatalf("second verification error = %v, want rejected", err)
	}
	if len(repo.updatedStatuses) != 2 || repo.updatedStatuses[0] != documents.StatusVerified || repo.updatedStatuses[1] != documents.StatusRejected {
		t.Fatalf("persisted statuses = %v", repo.updatedStatuses)
	}
}

func TestValidateForBookingRejectsForeignSavedDocument(t *testing.T) {
	repo := &fakeDocumentRepository{
		rule:   &documents.DocumentRule{AllowInternalPassport: true},
		getErr: documents.ErrNotFound,
	}
	svc := &Service{repo: repo, verificationSecret: []byte("secret"), now: time.Now}
	documentID := 77
	ownerID := 1
	_, err := svc.ValidateForBooking(context.Background(), validBookingInput(&ownerID, nil, &documentID))
	if !errors.Is(err, apierror.ErrNotFound) {
		t.Fatalf("ValidateForBooking() error = %v, want not found", err)
	}
}

func TestValidateForBookingRejectsExpiredDocument(t *testing.T) {
	repo := &fakeDocumentRepository{rule: &documents.DocumentRule{AllowInternalPassport: true}}
	svc := &Service{repo: repo, verificationSecret: []byte("secret"), now: time.Now}
	in := validBookingInput(nil, nil, nil)
	in.Document.Type = "internal_passport"
	in.Document.Number = "1234567890"
	in.Document.ExpiresAt = "2026-09-09"

	_, err := svc.ValidateForBooking(context.Background(), in)
	if !errors.Is(err, apierror.ErrDocumentExpired) {
		t.Fatalf("ValidateForBooking() error = %v, want expired", err)
	}
}

func TestCreateMapsRepositoryDomainErrors(t *testing.T) {
	tests := []struct {
		name    string
		repoErr error
		want    error
	}{
		{name: "owner not found", repoErr: documents.ErrNotFound, want: apierror.ErrNotFound},
		{name: "duplicate", repoErr: documents.ErrAlreadyExists, want: apierror.ErrDocumentAlreadyExists},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := &fakeDocumentRepository{createErr: test.repoErr}
			svc := &Service{repo: repo, verificationSecret: []byte("secret"), now: time.Now}
			_, err := svc.Create(context.Background(), 7, documents.SaveDocumentInput{
				Type: documents.TypeInternalPassport, Number: "1234567890",
			})
			if !errors.Is(err, test.want) {
				t.Fatalf("Create() error = %v, want %v", err, test.want)
			}
		})
	}
}

func validBookingInput(ownerID *int, passengerID *int, documentID *int) documents.BookingValidationInput {
	return documents.BookingValidationInput{
		OwnerUserID: ownerID, SavedPassengerID: passengerID,
		Passenger: documents.BookingPassenger{
			FirstName: "Иван", LastName: "Иванов", BirthDate: "1990-01-01", IsRussian: true,
		},
		Document:  documents.BookingDocument{ID: documentID},
		Transport: "avia", IsInternational: false,
		DepartureTime: time.Date(2026, time.September, 10, 10, 0, 0, 0, time.UTC),
	}
}

func findDocumentWithStatus(t *testing.T, secret []byte, checkedAt time.Time, want string) string {
	t.Helper()
	for i := 0; i < 100000; i++ {
		number := fmt.Sprintf("%010d", i)
		status, _ := verifyDocument(secret, "internal_passport", number, checkedAt)
		if status == want {
			return number
		}
	}
	t.Fatalf("cannot find document with status %s", want)
	return ""
}

func findDocumentTransition(t *testing.T, secret []byte, first time.Time, firstStatus string, second time.Time, secondStatus string) string {
	t.Helper()
	for i := 0; i < 100000; i++ {
		number := fmt.Sprintf("%010d", i)
		statusA, _ := verifyDocument(secret, "internal_passport", number, first)
		statusB, _ := verifyDocument(secret, "internal_passport", number, second)
		if statusA == firstStatus && statusB == secondStatus {
			return number
		}
	}
	t.Fatalf("cannot find document transition %s -> %s", firstStatus, secondStatus)
	return ""
}

type fakeDocumentRepository struct {
	rule              *documents.DocumentRule
	ruleErr           error
	doc               *documents.Document
	getErr            error
	createErr         error
	updatedDocumentID int
	updatedStatus     string
	updatedStatuses   []string
}

func (f *fakeDocumentRepository) FindRule(context.Context, string, bool, int, bool) (*documents.DocumentRule, error) {
	return f.rule, f.ruleErr
}
func (f *fakeDocumentRepository) ListForUser(context.Context, int) ([]documents.Document, error) {
	return nil, nil
}
func (f *fakeDocumentRepository) ListForPassenger(context.Context, int, int) ([]documents.Document, error) {
	return nil, nil
}
func (f *fakeDocumentRepository) CreateForUser(context.Context, int, documents.SaveDocumentInput, string, string, *time.Time, time.Time) (*documents.Document, error) {
	return nil, f.createErr
}
func (f *fakeDocumentRepository) CreateForPassenger(context.Context, int, int, documents.SaveDocumentInput, string, string, *time.Time, time.Time) (*documents.Document, error) {
	return nil, f.createErr
}
func (f *fakeDocumentRepository) GetOwned(context.Context, int, int, *int) (*documents.Document, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	return f.doc, nil
}
func (f *fakeDocumentRepository) Update(context.Context, int, int, documents.SaveDocumentInput, string, string, *time.Time, time.Time) (*documents.Document, error) {
	return nil, nil
}
func (f *fakeDocumentRepository) UpdateVerification(_ context.Context, documentID int, status string, _ time.Time) error {
	f.updatedDocumentID = documentID
	f.updatedStatus = status
	f.updatedStatuses = append(f.updatedStatuses, status)
	return nil
}
func (f *fakeDocumentRepository) Delete(context.Context, int, int) error { return nil }
