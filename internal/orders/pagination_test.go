package orders

import "testing"

func TestNormalizeListLimit(t *testing.T) {
	tests := []struct {
		input int
		want  int
	}{
		{input: -1, want: DefaultListLimit},
		{input: 0, want: DefaultListLimit},
		{input: 1, want: 1},
		{input: MaxListLimit, want: MaxListLimit},
		{input: MaxListLimit + 1, want: DefaultListLimit},
	}
	for _, test := range tests {
		if got := NormalizeListLimit(test.input); got != test.want {
			t.Errorf("NormalizeListLimit(%d) = %d, want %d", test.input, got, test.want)
		}
	}
}
