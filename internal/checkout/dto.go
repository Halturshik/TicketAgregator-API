package checkout

type PayInput struct {
	OrderID           int    `json:"order_id"`
	GuestPaymentToken string `json:"guest_payment_token,omitempty"`
}

type PayOutput struct {
	OrderID      int    `json:"order_id"`
	PaymentID    int    `json:"payment_id"`
	Status       string `json:"status"`
	Amount       int    `json:"amount"`
	BonusSpent   int    `json:"bonus_spent"`
	BonusEarned  int    `json:"bonus_earned"`
	BonusBalance *int   `json:"bonus_balance,omitempty"`
}
