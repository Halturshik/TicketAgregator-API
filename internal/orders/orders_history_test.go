package orders

import "testing"

func TestHistoryPageSizeIsFive(t *testing.T) {
	for _, limit := range []int{-1, 0, 6, 100} {
		if got := NormalizeListLimit(limit); got != 5 {
			t.Fatalf("NormalizeListLimit(%d) = %d, want 5", limit, got)
		}
	}
	if got := NormalizeListLimit(3); got != 3 {
		t.Fatalf("NormalizeListLimit(3) = %d, want 3", got)
	}
}
