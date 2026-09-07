package bookingaccess

import (
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/fare"
	"github.com/Halturshik/TicketAgregator-API/internal/orders"
)

type Booking struct {
	ID                int
	OrderNumber       string
	UserID            *int
	GuestEmail        string
	GuestPaymentToken string
	Status            string
	TotalPrice        int
	CurrentTotalPrice int
	BonusSpent        int
	BonusEarned       int
	CreatedAt         time.Time
	PassengerCount    int
	Tickets           []Ticket
}

type Ticket struct {
	ID           int
	Number       string
	Status       string
	Transport    string
	FareType     string
	RefundPolicy fare.RefundPolicy
	Price        int
	Passenger    orders.PassengerSnapshot
	Document     orders.DocumentSnapshot
	Segments     []Segment
}

type Segment struct {
	Order         int
	Carrier       string
	CarrierCode   string
	RouteNumber   string
	FromCity      string
	ToCity        string
	DepartureTime time.Time
	ArrivalTime   time.Time
}
