package refunds

import "time"

type Input struct {
	TicketIDs         []int  `json:"ticket_ids,omitempty"`
	All               bool   `json:"all,omitempty"`
	GuestPaymentToken string `json:"guest_payment_token,omitempty"`
	IdempotencyKey    string `json:"-"`
}

type QuoteOutput struct {
	OrderID       int         `json:"order_id"`
	Refundable    bool        `json:"refundable"`
	Reason        string      `json:"reason,omitempty"`
	CashAmount    int         `json:"cash_amount"`
	BonusRestored int         `json:"bonus_restored"`
	BonusRevoked  int         `json:"bonus_revoked"`
	Items         []QuoteItem `json:"items"`
}

type QuoteItem struct {
	TicketID          int    `json:"ticket_id"`
	SupplierCode      string `json:"supplier_code"`
	Refundable        bool   `json:"refundable"`
	Reason            string `json:"reason,omitempty"`
	RefundPercent     int    `json:"refund_percent"`
	GrossAmount       int    `json:"gross_amount"`
	GrossRefundAmount int    `json:"gross_refund_amount"`
}

type Result struct {
	RefundID         int         `json:"refund_id"`
	OrderID          int         `json:"order_id"`
	OrderStatus      string      `json:"order_status"`
	Status           string      `json:"status"`
	SupplierRefundID string      `json:"supplier_refund_id,omitempty"`
	FailureCode      string      `json:"failure_code,omitempty"`
	AttemptCount     int         `json:"attempt_count"`
	NextRetryAt      *time.Time  `json:"next_retry_at,omitempty"`
	CashAmount       int         `json:"cash_amount"`
	BonusRestored    int         `json:"bonus_restored"`
	BonusRevoked     int         `json:"bonus_revoked"`
	BonusBalance     *int        `json:"bonus_balance,omitempty"`
	BonusDebt        *int        `json:"bonus_debt,omitempty"`
	Items            []QuoteItem `json:"items"`
}
