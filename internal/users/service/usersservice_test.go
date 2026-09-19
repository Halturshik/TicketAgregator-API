package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/users"
)

func TestGetProfileMapsDomainNotFound(t *testing.T) {
	svc := NewService(&fakeUserRepository{getErr: users.ErrNotFound})
	_, err := svc.GetProfile(context.Background(), 7)
	if !errors.Is(err, apierror.ErrNotFound) {
		t.Fatalf("GetProfile() error = %v, want not found", err)
	}
}

func TestUpdateProfileValidatesBeforeRepository(t *testing.T) {
	repo := &fakeUserRepository{}
	svc := NewService(repo)
	_, err := svc.UpdateProfile(context.Background(), 7, users.UpdateProfileInput{BirthDate: "bad-date"})
	if err == nil {
		t.Fatal("UpdateProfile() expected validation error")
	}
	if repo.updateCalled {
		t.Fatal("repository was called for invalid profile")
	}
}

type fakeUserRepository struct {
	getErr       error
	updateErr    error
	updateCalled bool
}

func (f *fakeUserRepository) GetProfile(context.Context, int) (*users.Profile, error) {
	return nil, f.getErr
}

func (f *fakeUserRepository) UpdateProfile(context.Context, int, users.UpdateProfileInput, time.Time) (*users.Profile, error) {
	f.updateCalled = true
	return nil, f.updateErr
}
