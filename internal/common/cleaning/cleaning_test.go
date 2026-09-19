package cleaning

import "testing"

func TestEmailTrimsSpacesAndNormalizesCase(t *testing.T) {
	if got := Email("  Traveler@Example.COM "); got != "traveler@example.com" {
		t.Fatalf("Email() = %q, want traveler@example.com", got)
	}
}
