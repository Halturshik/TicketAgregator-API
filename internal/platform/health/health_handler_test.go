package health

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestLiveReturnsOKWithoutDependencyChecks(t *testing.T) {
	called := false
	handler := New(time.Second, Dependency{
		Name: "postgres",
		Probe: ProbeFunc(func(context.Context) error {
			called = true
			return errors.New("database unavailable")
		}),
	})
	recorder := httptest.NewRecorder()

	handler.Live(recorder, httptest.NewRequest(http.MethodGet, "/health/live", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if called {
		t.Fatal("liveness called a dependency probe")
	}
	if recorder.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("Cache-Control = %q", recorder.Header().Get("Cache-Control"))
	}
	if body := recorder.Body.String(); !strings.Contains(body, `"status":"alive"`) {
		t.Fatalf("body = %s", body)
	}
}

func TestReadyReturnsDependencyStatuses(t *testing.T) {
	tests := []struct {
		name       string
		probeError error
		wantStatus int
		wantBody   string
	}{
		{name: "ready", wantStatus: http.StatusOK, wantBody: `"status":"ready"`},
		{
			name:       "not ready",
			probeError: errors.New("postgres://user:secret@host/database"),
			wantStatus: http.StatusServiceUnavailable,
			wantBody:   `"status":"not_ready"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := New(time.Second, Dependency{
				Name: "postgres",
				Probe: ProbeFunc(func(context.Context) error {
					return tt.probeError
				}),
			})
			recorder := httptest.NewRecorder()

			handler.Ready(recorder, httptest.NewRequest(http.MethodGet, "/health/ready", nil))

			if recorder.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, tt.wantStatus)
			}
			body := recorder.Body.String()
			if !strings.Contains(body, tt.wantBody) {
				t.Fatalf("body = %s, want %s", body, tt.wantBody)
			}
			if !strings.Contains(body, `"postgres":"`) {
				t.Fatalf("body = %s, want postgres status", body)
			}
			if strings.Contains(body, "secret") {
				t.Fatalf("readiness response exposed internal error: %s", body)
			}
		})
	}
}
