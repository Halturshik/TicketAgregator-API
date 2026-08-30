package bonus

import "math"

const (
	EarnRate        = 0.02
	MaxSpendDivisor = 2

	TransactionTypeSpend = "spend"
	TransactionTypeEarn  = "earn"

	TransactionTypeRefundRestore = "refund_restore"
	TransactionTypeRefundRevoke  = "refund_revoke"
	TransactionTypeDebtCreate    = "debt_create"
	TransactionTypeDebtRepay     = "debt_repay"
)

func Earned(amount int) int {
	return int(math.Floor(float64(amount) * EarnRate))
}

func MaxSpend(total int) int {
	return total / MaxSpendDivisor
}
