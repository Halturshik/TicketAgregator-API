package bonus

import "testing"

func TestBonusRulesUseWholePointsAndHalfPriceLimit(t *testing.T) {
	tests := []struct {
		amount  int
		earned  int
		maximum int
	}{
		{amount: 0, earned: 0, maximum: 0},
		{amount: 49, earned: 0, maximum: 24},
		{amount: 50, earned: 1, maximum: 25},
		{amount: 5999, earned: 119, maximum: 2999},
		{amount: 6000, earned: 120, maximum: 3000},
	}
	for _, test := range tests {
		if got := Earned(test.amount); got != test.earned {
			t.Errorf("Earned(%d) = %d, want %d", test.amount, got, test.earned)
		}
		if got := MaxSpend(test.amount); got != test.maximum {
			t.Errorf("MaxSpend(%d) = %d, want %d", test.amount, got, test.maximum)
		}
	}
}
