package fare

import "time"

type RefundTier struct {
	MinHoursBeforeDeparture int `json:"min_hours_before_departure"`
	RefundPercent           int `json:"refund_percent"`
}

type RefundPolicy struct {
	Code       string       `json:"code"`
	Refundable bool         `json:"refundable"`
	Tiers      []RefundTier `json:"tiers,omitempty"`
}

func RefundPercent(policy RefundPolicy, departure time.Time, now time.Time) (int, bool) {
	if !policy.Refundable || !departure.After(now) {
		return 0, false
	}

	remaining := departure.Sub(now)
	percent := 0
	for _, tier := range policy.Tiers {
		if remaining >= time.Duration(tier.MinHoursBeforeDeparture)*time.Hour {
			percent = tier.RefundPercent
		}
	}
	return percent, percent > 0
}
