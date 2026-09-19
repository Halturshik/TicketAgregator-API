package logger

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Halturshik/TicketAgregator-API/internal/platform/requestid"
)

func TestHTTPMiddlewarePropagatesRequestIDAndLogsRequest(t *testing.T) {
	var output bytes.Buffer
	log, err := New(Options{Service: "test", Format: FormatJSON, Level: "info", Output: &output})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	previous := slog.Default()
	slog.SetDefault(log)
	defer slog.SetDefault(previous)

	handler := HTTPMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, ok := requestid.FromContext(r.Context())
		if !ok || id == "" {
			t.Fatal("request id is missing from context")
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("ok"))
	}))
	request := httptest.NewRequest(http.MethodPost, "/orders", nil)
	request.Header.Set(requestid.Header, "2b0d9d2b-0194-4dc6-804b-ec13ab1c6b40")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Header().Get(requestid.Header) != "2b0d9d2b-0194-4dc6-804b-ec13ab1c6b40" {
		t.Fatalf("response request id = %q", response.Header().Get(requestid.Header))
	}
	var entry map[string]any
	if err := json.Unmarshal(output.Bytes(), &entry); err != nil {
		t.Fatalf("decode log: %v; output = %q", err, output.String())
	}
	if entry["request_id"] != response.Header().Get(requestid.Header) {
		t.Fatalf("logged request_id = %v", entry["request_id"])
	}
	if entry["status"] != float64(http.StatusCreated) {
		t.Fatalf("status = %v", entry["status"])
	}
	if entry["response_bytes"] != float64(2) {
		t.Fatalf("response_bytes = %v", entry["response_bytes"])
	}
}

func TestHTTPMiddlewareDoesNotLogHealthChecks(t *testing.T) {
	var output bytes.Buffer
	log, err := New(Options{Service: "test", Format: FormatJSON, Level: "info", Output: &output})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	previous := slog.Default()
	slog.SetDefault(log)
	defer slog.SetDefault(previous)

	handler := HTTPMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/health/ready", nil))
	if output.Len() != 0 {
		t.Fatalf("health request was logged: %q", output.String())
	}
}
