package fare

const (
	NonRefundable = "non_refundable"
	Standard      = "standard"
	Flexible      = "flexible"

	StandardRefundHours  = 72
	FlexibleRefundHours  = 24
	FlexiblePremiumHours = 72

	PartialRefundPercent = 70
	PremiumRefundPercent = 90
)

var policies = map[string]RefundPolicy{
	NonRefundable: {
		Code:       NonRefundable,
		Refundable: false,
	},
	Standard: {
		Code:       Standard,
		Refundable: true,
		Tiers: []RefundTier{
			{MinHoursBeforeDeparture: StandardRefundHours, RefundPercent: PremiumRefundPercent},
		},
	},
	Flexible: {
		Code:       Flexible,
		Refundable: true,
		Tiers: []RefundTier{
			{MinHoursBeforeDeparture: FlexibleRefundHours, RefundPercent: PartialRefundPercent},
			{MinHoursBeforeDeparture: FlexiblePremiumHours, RefundPercent: PremiumRefundPercent},
		},
	},
}

func Policy(code string) (RefundPolicy, bool) {
	policy, ok := policies[code]
	if !ok {
		return RefundPolicy{}, false
	}
	policy.Tiers = append([]RefundTier(nil), policy.Tiers...)
	return policy, true
}
