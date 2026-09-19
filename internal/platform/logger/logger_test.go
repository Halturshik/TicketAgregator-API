package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"github.com/Halturshik/TicketAgregator-API/internal/platform/requestid"
)

func TestNewJSONAddsServiceAndRequestID(t *testing.T) {
	var output bytes.Buffer
	log, err := New(Options{
		Service: "test-service",
		Format:  FormatJSON,
		Level:   "info",
		Output:  &output,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	ctx := requestid.WithContext(context.Background(), "request-123")
	log.InfoContext(ctx, "Событие", slog.Int("order_id", 42))

	var entry map[string]any
	if err := json.Unmarshal(output.Bytes(), &entry); err != nil {
		t.Fatalf("decode log: %v; output = %q", err, output.String())
	}
	if entry["service"] != "test-service" {
		t.Fatalf("service = %v", entry["service"])
	}
	if entry["request_id"] != "request-123" {
		t.Fatalf("request_id = %v", entry["request_id"])
	}
	if entry["order_id"] != float64(42) {
		t.Fatalf("order_id = %v", entry["order_id"])
	}
}

func TestNewRespectsLevel(t *testing.T) {
	var output bytes.Buffer
	log, err := New(Options{Service: "test", Format: FormatText, Level: "warn", Output: &output})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	log.Info("hidden")
	log.Warn("visible")
	if strings.Contains(output.String(), "hidden") || !strings.Contains(output.String(), "visible") {
		t.Fatalf("unexpected output: %q", output.String())
	}
}

func TestNewRejectsInvalidOptions(t *testing.T) {
	tests := []Options{
		{Format: FormatText, Level: "info"},
		{Service: "test", Format: "xml", Level: "info"},
		{Service: "test", Format: FormatText, Level: "verbose"},
	}
	for _, options := range tests {
		if _, err := New(options); err == nil {
			t.Fatalf("New(%+v) returned nil error", options)
		}
	}
}

func TestMaskEmail(t *testing.T) {
	tests := map[string]string{
		"user@example.com": "u***@example.com",
		"a@example.com":    "a@example.com",
		"иван@example.com": "и***@example.com",
		"invalid":          "***",
		"@example.com":     "***",
	}
	for input, expected := range tests {
		if actual := MaskEmail(input); actual != expected {
			t.Fatalf("MaskEmail(%q) = %q, want %q", input, actual, expected)
		}
	}
}
