package orders

import "github.com/Halturshik/TicketAgregator-API/internal/fare"

type Order struct {
	ID                int      `json:"id"`
	OrderNumber       string   `json:"order_number"`
	UserID            *int     `json:"user_id,omitempty"`
	GuestEmail        string   `json:"guest_email,omitempty"`
	GuestPaymentToken string   `json:"guest_payment_token,omitempty"`
	Status            string   `json:"status"`
	TotalPrice        int      `json:"total_price"`
	CurrentTotalPrice int      `json:"current_total_price"`
	BonusSpent        int      `json:"bonus_spent"`
	BonusEarned       int      `json:"bonus_earned"`
	PayableAmount     int      `json:"payable_amount"`
	ExpiresAt         string   `json:"expires_at"`
	Tickets           []Ticket `json:"tickets,omitempty"`
}

type Ticket struct {
	ID                  int                     `json:"id"`
	TicketNumber        string                  `json:"ticket_number"`
	SupplierCode        string                  `json:"supplier_code"`
	SupplierOfferID     string                  `json:"supplier_offer_id"`
	FareType            string                  `json:"fare_type"`
	RefundPolicy        fare.RefundPolicy       `json:"refund_policy"`
	RefundPolicyVersion int                     `json:"refund_policy_version"`
	Transport           string                  `json:"transport"`
	IsInternational     bool                    `json:"is_international"`
	Price               int                     `json:"price"`
	Status              string                  `json:"status"`
	Passenger           PassengerSnapshot       `json:"passenger"`
	Document            DocumentSnapshot        `json:"document"`
	Segments            []TicketSegmentSnapshot `json:"segments"`
}
