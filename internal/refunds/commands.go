package refunds

import "time"

type CreateParams struct {
	OrderID                int
	PaymentID              int
	IdempotencyKey         string
	RequestHash            string
	SupplierCode           string
	NextRetryAt            time.Time
	ReconciliationDeadline time.Time
	CashAmount             int
	BonusRestored          int
	BonusRevoked           int
	Items                  []CreateItemParams
}

type CreateItemParams struct {
	TicketID             int
	SupplierCode         string
	TicketNumber         string
	SupplierOfferID      string
	FareType             string
	DepartureAt          time.Time
	Reason               string
	RefundPercent        int
	GrossAmount          int
	SupplierRefundAmount int
}

type SuccessParams struct {
	RefundID          int
	SupplierRefundID  string
	CashAmount        int
	BonusRestored     int
	BonusRevoked      int
	BonusBalanceAfter *int
	BonusDebtAfter    *int
	Items             []SuccessItemParams
}

type SuccessItemParams struct {
	TicketID             int
	Reason               string
	RefundPercent        int
	SupplierRefundAmount int
}

type OrderRefundParams struct {
	OrderID           int
	Status            string
	CurrentTotalPrice int
	BonusSpent        int
	BonusEarned       int
	PayableAmount     int
}

type FailureParams struct {
	RefundID         int
	SupplierRefundID string
	FailureCode      string
	Items            []FailureItemParams
}

type FailureItemParams struct {
	TicketID int
	Reason   string
}

type BonusParams struct {
	UserID        int
	OrderID       int
	RefundID      int
	RestoreAmount int
	RevokeAmount  int
}

type BonusResult struct {
	Balance     int
	Debt        int
	DebtCreated int
}

type AttemptErrorParams struct {
	RefundID    int
	Message     string
	Status      string
	FailureCode string
	NextRetryAt time.Time
}
