package orders

import (
	"time"
)

type PaymentData struct {
	ID            int
	UserID        *int
	GuestToken    string
	Status        string
	PayableAmount int
	BonusSpent    int
	BonusEarned   int
	ExpiresAt     time.Time
}
