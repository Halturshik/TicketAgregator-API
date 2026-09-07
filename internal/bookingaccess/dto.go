package bookingaccess

import "time"

type PublicSummary struct {
	Kind           string        `json:"kind"`
	OrderNumber    string        `json:"order_number"`
	Transport      string        `json:"transport"`
	TicketCount    int           `json:"ticket_count"`
	PassengerCount int           `json:"passenger_count"`
	Directions     []Direction   `json:"directions"`
	Ticket         *PublicTicket `json:"ticket,omitempty"`
}

type PublicTicket struct {
	TicketNumber string          `json:"ticket_number"`
	Passenger    string          `json:"passenger"`
	Segments     []PublicSegment `json:"segments"`
}

type Direction struct {
	FromCity      string `json:"from_city"`
	ToCity        string `json:"to_city"`
	DepartureTime string `json:"departure_time"`
	ArrivalTime   string `json:"arrival_time"`
	TransferCount int    `json:"transfer_count"`
}

type PublicSegment struct {
	Carrier       string `json:"carrier"`
	CarrierCode   string `json:"carrier_code"`
	RouteNumber   string `json:"route_number"`
	FromCity      string `json:"from_city"`
	ToCity        string `json:"to_city"`
	DepartureTime string `json:"departure_time"`
	ArrivalTime   string `json:"arrival_time"`
}

type AccessRequestOutput struct {
	ChallengeID string `json:"challenge_id"`
	Message     string `json:"message"`
}

type AccessOutput struct {
	OrderNumber string    `json:"order_number,omitempty"`
	ExpiresAt   time.Time `json:"expires_at"`
	Token       string    `json:"-"`
}

type BookingDetails struct {
	OrderNumber       string         `json:"order_number"`
	Status            string         `json:"status"`
	CurrentTotalPrice int            `json:"current_total_price"`
	BonusSpent        int            `json:"bonus_spent"`
	BonusEarned       int            `json:"bonus_earned"`
	CreatedAt         string         `json:"created_at"`
	PassengerCount    int            `json:"passenger_count"`
	Directions        []Direction    `json:"directions"`
	Tickets           []DetailTicket `json:"tickets"`
}

type DetailTicket struct {
	TicketNumber string          `json:"ticket_number"`
	Status       string          `json:"status"`
	Transport    string          `json:"transport"`
	FareType     string          `json:"fare_type"`
	Passenger    string          `json:"passenger"`
	DocumentType string          `json:"document_type"`
	Document     string          `json:"document"`
	Segments     []PublicSegment `json:"segments"`
}

type RefundQuote struct {
	OrderNumber   string            `json:"order_number"`
	Refundable    bool              `json:"refundable"`
	Reason        string            `json:"reason,omitempty"`
	CashAmount    int               `json:"cash_amount"`
	BonusRestored int               `json:"bonus_restored"`
	BonusRevoked  int               `json:"bonus_revoked"`
	Items         []RefundQuoteItem `json:"items"`
}

type RefundQuoteItem struct {
	TicketNumber      string `json:"ticket_number"`
	Refundable        bool   `json:"refundable"`
	Reason            string `json:"reason,omitempty"`
	RefundPercent     int    `json:"refund_percent"`
	GrossAmount       int    `json:"gross_amount"`
	GrossRefundAmount int    `json:"gross_refund_amount"`
}

type RefundResult struct {
	OrderNumber      string            `json:"order_number"`
	OrderStatus      string            `json:"order_status"`
	Status           string            `json:"status"`
	SupplierRefundID string            `json:"supplier_refund_id,omitempty"`
	FailureCode      string            `json:"failure_code,omitempty"`
	AttemptCount     int               `json:"attempt_count"`
	NextRetryAt      *time.Time        `json:"next_retry_at,omitempty"`
	CashAmount       int               `json:"cash_amount"`
	BonusRestored    int               `json:"bonus_restored"`
	BonusRevoked     int               `json:"bonus_revoked"`
	BonusBalance     *int              `json:"bonus_balance,omitempty"`
	BonusDebt        *int              `json:"bonus_debt,omitempty"`
	Items            []RefundQuoteItem `json:"items"`
}
