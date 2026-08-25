package orders

import "time"

const (
	OrderStatusCreated = "created"
	OrderStatusPaid    = "paid"
	OrderStatusExpired = "expired"

	TicketStatusBooked  = "booked"
	TicketStatusPaid    = "paid"
	TicketStatusExpired = "expired"

	OrderTTL = 15 * time.Minute
)
