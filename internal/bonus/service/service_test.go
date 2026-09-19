package service

import (
	"context"
	"testing"

	"github.com/Halturshik/TicketAgregator-API/internal/bonus"
)

func TestListNormalizesPaginationBeforeRepository(t *testing.T) {
	repo := &fakeBonusRepository{}
	svc := NewService(repo)

	if _, err := svc.List(context.Background(), 7, 100, -5); err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if repo.userID != 7 || repo.limit != 10 || repo.offset != 0 {
		t.Fatalf("repository input = user:%d limit:%d offset:%d", repo.userID, repo.limit, repo.offset)
	}
}

type fakeBonusRepository struct {
	userID int
	limit  int
	offset int
}

func (f *fakeBonusRepository) List(_ context.Context, userID int, limit int, offset int) ([]bonus.Transaction, error) {
	f.userID = userID
	f.limit = limit
	f.offset = offset
	return nil, nil
}
