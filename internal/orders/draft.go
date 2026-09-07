package orders

import (
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/fare"
)

type CreateOrderParams struct {
	OrderNumber       string
	UserID            *int
	GuestEmail        string
	GuestPaymentToken string
	TotalPrice        int
	CurrentTotalPrice int
	BonusSpent        int
	BonusEarned       int
	PayableAmount     int
	ExpiresAt         time.Time
	Passengers        []OrderPassengerDraft
	Tickets           []TicketDraft
}

type OrderPassengerDraft struct {
	Source              string
	SavedPassengerID    *int
	SavedDocumentID     *int
	SaveChanges         bool
	Passenger           PassengerSnapshot
	Document            DocumentSnapshot
	DocumentFingerprint string
}

type TicketDraft struct {
	TicketNumber        string
	PassengerIndex      int
	SupplierCode        string
	SupplierOfferID     string
	FareType            string
	RefundPolicy        fare.RefundPolicy
	RefundPolicyVersion int
	Transport           string
	IsInternational     bool
	Price               int
	Segments            []TicketSegmentSnapshot
}
