package supplier

import "github.com/Halturshik/TicketAgregator-API/internal/fare"

type RefundOperation struct {
	ID             string
	ProviderCode   string
	IdempotencyKey string
	RequestHash    string
	Status         string
	FailureCode    string
	Items          []RefundOperationItem
}

type RefundOperationItem struct {
	TicketID        int
	TicketNumber    string
	SupplierOfferID string
	FareType        string
	DepartureUnix   int64
	GrossAmount     int
	Refunded        bool
	Reason          string
	RefundPercent   int
	RefundAmount    int
	Policy          fare.RefundPolicy
}
