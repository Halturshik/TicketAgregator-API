package orders

import "github.com/Halturshik/TicketAgregator-API/internal/search"

type PassengerSnapshot struct {
	SavedPassengerID *int   `json:"saved_passenger_id,omitempty"`
	FirstName        string `json:"first_name"`
	MiddleName       string `json:"middle_name,omitempty"`
	LastName         string `json:"last_name"`
	BirthDate        string `json:"birth_date"`
	IsRussian        bool   `json:"is_russian"`
}

type DocumentSnapshot struct {
	Type   string `json:"type"`
	Number string `json:"number"`
}

type PassengerBooking struct {
	Passenger PassengerSnapshot `json:"passenger"`
	Document  DocumentSnapshot  `json:"document"`
}

type CreateOrderInput struct {
	SearchID   string             `json:"search_id"`
	OfferIDs   []string           `json:"offer_ids"`
	GuestEmail string             `json:"guest_email,omitempty"`
	UseBonus   int                `json:"use_bonus,omitempty"`
	Passengers []PassengerBooking `json:"passengers"`
}

type Order struct {
	ID                int      `json:"id"`
	UserID            *int     `json:"user_id,omitempty"`
	GuestEmail        string   `json:"guest_email,omitempty"`
	GuestPaymentToken string   `json:"guest_payment_token,omitempty"`
	Status            string   `json:"status"`
	TotalPrice        int      `json:"total_price"`
	BonusSpent        int      `json:"bonus_spent"`
	BonusEarned       int      `json:"bonus_earned"`
	PayableAmount     int      `json:"payable_amount"`
	Tickets           []Ticket `json:"tickets,omitempty"`
}

type Ticket struct {
	ID              int               `json:"id"`
	TicketNumber    string            `json:"ticket_number"`
	Transport       string            `json:"transport"`
	IsInternational bool              `json:"is_international"`
	Price           int               `json:"price"`
	Status          string            `json:"status"`
	Passenger       PassengerSnapshot `json:"passenger"`
	Document        DocumentSnapshot  `json:"document"`
	Segments        []search.Segment  `json:"segments"`
}

type CreateOrderParams struct {
	UserID            *int
	GuestEmail        string
	GuestPaymentToken string
	TotalPrice        int
	BonusSpent        int
	BonusEarned       int
	PayableAmount     int
	Tickets           []TicketDraft
}

type TicketDraft struct {
	TicketNumber    string
	Transport       string
	IsInternational bool
	Price           int
	Passenger       PassengerSnapshot
	Document        DocumentSnapshot
	Segments        []search.Segment
}

type ListFilter struct {
	Transport string
	Limit     int
	Offset    int
}
