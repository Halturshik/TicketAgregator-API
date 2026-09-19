package number

import (
	"regexp"
	"testing"
)

func TestOrderNumberFormat(t *testing.T) {
	pattern := regexp.MustCompile(`^[A-Z]{3}-[0-9]{5}$`)
	seen := make(map[string]struct{})
	for range 100 {
		value, err := NewOrderNumber()
		if err != nil {
			t.Fatalf("NewOrderNumber() error = %v", err)
		}
		if !pattern.MatchString(value) {
			t.Fatalf("order number %q does not match %s", value, pattern)
		}
		if _, exists := seen[value]; exists {
			t.Fatalf("duplicate order number %q", value)
		}
		seen[value] = struct{}{}
	}
}

func TestSeriesFormatsAndSharedSegments(t *testing.T) {
	tests := []struct {
		transport string
		pattern   *regexp.Regexp
		shared    func(string) string
	}{
		{transport: "avia", pattern: regexp.MustCompile(`^[A-Z]{2}-[0-9]{8}$`), shared: func(value string) string { return value[:6] }},
		{transport: "rail", pattern: regexp.MustCompile(`^[0-9][A-Z][0-9]{2}[A-Z][0-9][A-Z]$`), shared: func(value string) string { return value[:3] }},
		{transport: "bus", pattern: regexp.MustCompile(`^[0-9]{6}[A-Z]$`), shared: func(value string) string { return value[len(value)-2:] }},
	}
	for _, tt := range tests {
		t.Run(tt.transport, func(t *testing.T) {
			series, err := NewSeries(tt.transport)
			if err != nil {
				t.Fatalf("NewSeries() error = %v", err)
			}
			seen := make(map[string]struct{})
			shared := ""
			for i := 0; i < 12; i++ {
				value, err := series.Next()
				if err != nil {
					t.Fatalf("Next() error = %v", err)
				}
				if !tt.pattern.MatchString(value) {
					t.Fatalf("ticket number %q does not match %s", value, tt.pattern)
				}
				if _, exists := seen[value]; exists {
					t.Fatalf("duplicate ticket number %q", value)
				}
				seen[value] = struct{}{}
				if i == 0 {
					shared = tt.shared(value)
				} else if tt.shared(value) != shared {
					t.Fatalf("ticket number %q does not share series segment %q", value, shared)
				}
			}
		})
	}
}

func TestNewSeriesRejectsUnknownTransport(t *testing.T) {
	if _, err := NewSeries("boat"); err == nil {
		t.Fatal("NewSeries() accepted unsupported transport")
	}
}
