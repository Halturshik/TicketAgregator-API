package refunds

import (
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/fare"
)

type OrderData struct {
	ID         int
	UserID     *int
	GuestToken string
	Status     string
}

type TicketData struct {
	ID                  int
	TicketNumber        string
	Status              string
	SupplierCode        string
	SupplierOfferID     string
	FareType            string
	RefundPolicy        fare.RefundPolicy
	RefundPolicyVersion int
	Price               int
	BonusSpent          int
	BonusEarned         int
	PayableAmount       int
	DepartureAt         time.Time
}

type Operation struct {
	ID                     int
	Order                  OrderData
	PaymentID              int
	Status                 string
	IdempotencyKey         string
	RequestHash            string
	SupplierCode           string
	SupplierRefundID       string
	FailureCode            string
	LastError              string
	AttemptCount           int
	NextRetryAt            time.Time
	ReconciliationDeadline time.Time
	CashAmount             int
	BonusRestored          int
	BonusRevoked           int
	BonusBalanceAfter      *int
	BonusDebtAfter         *int
	Items                  []OperationItem
}

type OperationItem struct {
	TicketData
	Reason        string
	RefundPercent int
	GrossAmount   int
	CashRefunded  int
	BonusRestored int
	BonusRevoked  int
}
