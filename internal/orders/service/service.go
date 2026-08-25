package service

import (
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/bonus"
	"github.com/Halturshik/TicketAgregator-API/internal/orders"
)

type Service struct {
	repo       orders.Repository
	search     orders.SearchReader
	documents  orders.DocumentValidator
	passengers orders.PassengerReader
	bonus      bonus.BalanceReader
	now        func() time.Time
}

func NewService(
	repo orders.Repository,
	search orders.SearchReader,
	documents orders.DocumentValidator,
	passengers orders.PassengerReader,
	bonusBalance bonus.BalanceReader,
) orders.Service {
	return &Service{
		repo:       repo,
		search:     search,
		documents:  documents,
		passengers: passengers,
		bonus:      bonusBalance,
		now:        time.Now,
	}
}
