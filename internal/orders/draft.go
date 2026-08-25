package orders

import "time"

type CreateOrderParams struct {
	UserID            *int
	GuestEmail        string
	GuestPaymentToken string
	TotalPrice        int
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
	TicketNumber    string
	PassengerIndex  int
	Transport       string
	IsInternational bool
	Price           int
	Segments        []TicketSegmentSnapshot
}
