package ratelimit_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Halturshik/TicketAgregator-API/internal/app"
	"github.com/Halturshik/TicketAgregator-API/internal/auth"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/ratelimit"
)

func TestSearchRateLimitsGuestAndUserSeparately(t *testing.T) {
	_, client := testRedis(t)
	middleware := app.NewRateLimitMiddleware(ratelimit.New(client))
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })

	assertAllowedRequests(t, middleware.Search(next), 10, false)
	assertAllowedRequests(t, middleware.Search(next), 20, true)
}

func TestPageAndPublicLookupRateLimits(t *testing.T) {
	_, client := testRedis(t)
	middleware := app.NewRateLimitMiddleware(ratelimit.New(client))
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })

	assertAllowedRequests(t, middleware.SearchPage(next), 60, false)
	assertAllowedRequests(t, middleware.PublicLookup(next), 20, false)
}

func assertAllowedRequests(t *testing.T, handler http.Handler, limit int, authenticated bool) {
	t.Helper()
	for requestNumber := 1; requestNumber <= limit+1; requestNumber++ {
		request := httptest.NewRequest(http.MethodPost, "/", nil)
		request.RemoteAddr = "192.0.2.10:3000"
		if authenticated {
			request = request.WithContext(auth.WithUserID(request.Context(), 42))
		}
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		expected := http.StatusNoContent
		if requestNumber > limit {
			expected = http.StatusTooManyRequests
		}
		if response.Code != expected {
			t.Fatalf("request %d/%d: expected status %d, got %d", requestNumber, limit+1, expected, response.Code)
		}
		if requestNumber > limit && response.Header().Get("Retry-After") == "" {
			t.Fatal("rate limited response has no Retry-After header")
		}
	}
}
