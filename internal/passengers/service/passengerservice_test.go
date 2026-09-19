package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/passengers"
)

func TestValidPassengerBirthDateAllowsChildren(t *testing.T) {
	if _, err := validPassengerBirthDate("2022-01-02"); err != nil {
		t.Fatalf("expected child passenger birth date to be valid: %v", err)
	}
}

func TestValidPassengerBirthDateRejectsBadFormat(t *testing.T) {
	if _, err := validPassengerBirthDate("02.01.2022"); err == nil {
		t.Fatalf("expected non-ISO birth date to be invalid")
	}
}

func TestGetOwnedMapsDomainNotFound(t *testing.T) {
	svc := NewService(&fakePassengerRepository{getErr: passengers.ErrNotFound})
	_, err := svc.GetOwned(context.Background(), 7, 11)
	if !errors.Is(err, apierror.ErrNotFound) {
		t.Fatalf("GetOwned() error = %v, want not found", err)
	}
}

func TestUpdateValidatesAndNormalizesBeforeRepository(t *testing.T) {
	repo := &fakePassengerRepository{updated: &passengers.Passenger{ID: 11}}
	svc := NewService(repo)
	_, err := svc.Update(context.Background(), 7, 11, passengers.SavePassengerInput{
		FirstName: "  иВАН ", LastName: " иВАНОВ ", BirthDate: "1990-01-01", IsRussian: true,
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if repo.input.FirstName != "иВАН" || repo.input.LastName != "иВАНОВ" {
		t.Fatalf("normalized input = %+v", repo.input)
	}
}

func TestDeleteMapsWrappedDomainNotFound(t *testing.T) {
	svc := NewService(&fakePassengerRepository{deleteErr: errors.New("storage: " + passengers.ErrNotFound.Error())})
	if err := svc.Delete(context.Background(), 7, 11); errors.Is(err, apierror.ErrNotFound) {
		t.Fatal("plain text error must not be treated as domain not found")
	}

	svc = NewService(&fakePassengerRepository{deleteErr: wrappedPassengerNotFound{}})
	if err := svc.Delete(context.Background(), 7, 11); !errors.Is(err, apierror.ErrNotFound) {
		t.Fatalf("Delete() error = %v, want not found", err)
	}
}

type wrappedPassengerNotFound struct{}

func (wrappedPassengerNotFound) Error() string { return "wrapped passenger not found" }
func (wrappedPassengerNotFound) Unwrap() error { return passengers.ErrNotFound }

type fakePassengerRepository struct {
	getErr    error
	updateErr error
	deleteErr error
	updated   *passengers.Passenger
	input     passengers.SavePassengerInput
}

func (f *fakePassengerRepository) List(context.Context, int) ([]passengers.Passenger, error) {
	return nil, nil
}

func (f *fakePassengerRepository) GetOwned(context.Context, int, int) (*passengers.Passenger, error) {
	return nil, f.getErr
}

func (f *fakePassengerRepository) Update(_ context.Context, _ int, _ int, in passengers.SavePassengerInput, _ time.Time) (*passengers.Passenger, error) {
	f.input = in
	return f.updated, f.updateErr
}

func (f *fakePassengerRepository) Delete(context.Context, int, int) error {
	return f.deleteErr
}
