package bonus

import "math"

const (
	EarnRate        = 0.02
	MaxSpendDivisor = 2

	TransactionTypeSpend = "spend"
	TransactionTypeEarn  = "earn"
)

func Earned(amount int) int {
	return int(math.Floor(float64(amount) * EarnRate))
}

func MaxSpend(total int) int {
	return total / MaxSpendDivisor
}
