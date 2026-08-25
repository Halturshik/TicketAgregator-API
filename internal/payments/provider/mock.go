package provider

import (
	"context"
	"crypto/rand"

	"github.com/Halturshik/TicketAgregator-API/internal/payments"
)

type Mock struct{}

var _ payments.Provider = Mock{}

func NewMock() Mock {
	return Mock{}
}

func (Mock) Process(context.Context) (bool, error) {
	var value [1]byte
	if _, err := rand.Read(value[:]); err != nil {
		return false, err
	}
	return int(value[0])%probabilityScale >= failureRate, nil
}
