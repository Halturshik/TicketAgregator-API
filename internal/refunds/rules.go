package refunds

import "time"

const (
	MaxReconciliationAttempts = 8
	ReconciliationWindow      = 24 * time.Hour
)

var reconciliationRetryDelays = [...]time.Duration{
	15 * time.Second,
	1 * time.Minute,
	5 * time.Minute,
	15 * time.Minute,
	1 * time.Hour,
	3 * time.Hour,
	6 * time.Hour,
}

func ReconciliationRetryDelay(failedAttempts int) time.Duration {
	if failedAttempts <= 0 {
		return 0
	}
	index := failedAttempts - 1
	if index >= len(reconciliationRetryDelays) {
		return reconciliationRetryDelays[len(reconciliationRetryDelays)-1]
	}
	return reconciliationRetryDelays[index]
}
