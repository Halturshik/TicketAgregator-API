package service

import (
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/checkout"
	"github.com/Halturshik/TicketAgregator-API/internal/passengers"
	"github.com/Halturshik/TicketAgregator-API/internal/payments"
)

type Service struct {
	repo       checkout.Repository
	provider   payments.Provider
	passengers passengers.PaymentSynchronizer
	now        func() time.Time
}

func NewService(
	repo checkout.Repository,
	provider payments.Provider,
	passengerSynchronizer passengers.PaymentSynchronizer,
) checkout.Service {
	return &Service{
		repo:       repo,
		provider:   provider,
		passengers: passengerSynchronizer,
		now:        time.Now,
	}
}
