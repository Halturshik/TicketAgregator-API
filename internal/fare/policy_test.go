package fare

import (
	"testing"
	"time"
)

func TestRefundPercentHonorsExactBoundaries(t *testing.T) {
	now := time.Date(2026, time.August, 29, 12, 0, 0, 0, time.UTC)
	standard, _ := Policy(Standard)
	flexible, _ := Policy(Flexible)
	nonRefundable, _ := Policy(NonRefundable)

	tests := []struct {
		name       string
		policy     RefundPolicy
		departure  time.Time
		want       int
		refundable bool
	}{
		{name: "non refundable", policy: nonRefundable, departure: now.Add(30 * 24 * time.Hour)},
		{name: "flexible one second early", policy: flexible, departure: now.Add(24*time.Hour - time.Second)},
		{name: "flexible exactly 24 hours", policy: flexible, departure: now.Add(24 * time.Hour), want: 70, refundable: true},
		{name: "flexible before premium", policy: flexible, departure: now.Add(72*time.Hour - time.Second), want: 70, refundable: true},
		{name: "flexible exactly 72 hours", policy: flexible, departure: now.Add(72 * time.Hour), want: 90, refundable: true},
		{name: "standard before deadline", policy: standard, departure: now.Add(72*time.Hour - time.Second)},
		{name: "standard exactly 72 hours", policy: standard, departure: now.Add(72 * time.Hour), want: 90, refundable: true},
		{name: "already departed", policy: flexible, departure: now.Add(-time.Second)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, ok := RefundPercent(test.policy, test.departure, now)
			if got != test.want || ok != test.refundable {
				t.Fatalf("RefundPercent() = (%d, %t), want (%d, %t)", got, ok, test.want, test.refundable)
			}
		})
	}
}

func TestPolicyReturnsIndependentTierSlice(t *testing.T) {
	first, _ := Policy(Flexible)
	first.Tiers[0].RefundPercent = 1
	second, _ := Policy(Flexible)
	if second.Tiers[0].RefundPercent != PartialRefundPercent {
		t.Fatalf("shared policy state was mutated: %d", second.Tiers[0].RefundPercent)
	}
}
