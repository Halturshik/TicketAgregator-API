package service

import (
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/search"
	"github.com/Halturshik/TicketAgregator-API/internal/supplier"
	"github.com/google/uuid"
)

type Service struct {
	repo          search.Repository
	store         search.Store
	supplier      supplier.Gateway
	providerCodes []string
	carriers      []search.CarrierConfig
	now           func() time.Time
	newID         func() string
}

func NewService(
	repo search.Repository,
	store search.Store,
	supplierGateway supplier.Gateway,
	providerCodes []string,
	carriers []search.CarrierConfig,
) (search.Service, error) {
	if supplierGateway == nil || !validProviderCodes(providerCodes) || !validCarrierConfiguration(carriers) {
		return nil, search.ErrInvalidSupplierConfiguration
	}
	return &Service{
		repo:          repo,
		store:         store,
		supplier:      supplierGateway,
		providerCodes: append([]string(nil), providerCodes...),
		carriers:      append([]search.CarrierConfig(nil), carriers...),
		now:           time.Now,
		newID:         uuid.NewString,
	}, nil
}
