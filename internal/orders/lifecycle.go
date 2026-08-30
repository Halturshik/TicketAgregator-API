package orders

import "time"

const (
	OrderStatusCreated           = "created"
	OrderStatusPaid              = "paid"
	OrderStatusExpired           = "expired"
	OrderStatusPartiallyRefunded = "partially_refunded"
	OrderStatusRefunded          = "refunded"

	TicketStatusBooked        = "booked"
	TicketStatusPaid          = "paid"
	TicketStatusExpired       = "expired"
	TicketStatusRefundPending = "refund_pending"
	TicketStatusRefunded      = "refunded"

	OrderTTL = 15 * time.Minute
)
