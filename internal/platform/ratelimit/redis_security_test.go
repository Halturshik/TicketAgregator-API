package ratelimit_test

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/Halturshik/TicketAgregator-API/internal/bookingaccess"
	bookingstore "github.com/Halturshik/TicketAgregator-API/internal/bookingaccess/store"
	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/codegen"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/ratelimit"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestSharedVerificationCodeGeneratorPreservesAuthFormat(t *testing.T) {
	generator := codegen.NewNumericGenerator(codegen.DefaultNumericLength)
	for index := 0; index < 100; index++ {
		code, err := generator.GenerateVerificationCode()
		if err != nil {
			t.Fatalf("generate code: %v", err)
		}
		if !regexp.MustCompile(`^[0-9]{4}$`).MatchString(code) {
			t.Fatalf("unexpected verification code %q", code)
		}
	}
}

func TestBookingChallengeRateAttemptsAndCooldown(t *testing.T) {
	server, client := testRedis(t)
	store := bookingstore.NewChallengeStore(client)
	challenge := bookingaccess.Challenge{
		ID:      "6f745a7f-b0cd-443e-934d-b3aa0d9e542c",
		OrderID: 15, OrderNumber: "AHS-74539", Locator: "AHS-74539", Email: "guest@example.com",
	}
	ctx := context.Background()
	if err := store.Request(ctx, challenge, "1234"); err != nil {
		t.Fatalf("request code: %v", err)
	}
	if err := store.Request(ctx, challenge, "1234"); !errors.Is(err, apierror.ErrCodeRateLimited) {
		t.Fatalf("expected request rate limit, got %v", err)
	}

	for attempt := 1; attempt <= bookingaccess.MaxVerificationAttempts; attempt++ {
		_, err := store.Verify(ctx, challenge.ID, "0000")
		expected := error(apierror.ErrInvalidVerificationCode)
		if attempt == bookingaccess.MaxVerificationAttempts {
			expected = apierror.ErrTooManyAttempts
		}
		if !errors.Is(err, expected) {
			t.Fatalf("attempt %d: expected %v, got %v", attempt, expected, err)
		}
	}

	server.FastForward(bookingaccess.CodeRequestInterval)
	if err := store.Request(ctx, challenge, "1234"); !errors.Is(err, apierror.ErrTooManyAttempts) {
		t.Fatalf("expected cooldown after failed attempts, got %v", err)
	}
}

func TestBookingAccessTokenIsHashedAndOrderScoped(t *testing.T) {
	server, client := testRedis(t)
	store := bookingstore.NewAccessStore(client)
	ctx := context.Background()
	token, err := store.Issue(ctx, 77)
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}
	for _, key := range server.Keys() {
		value, _ := server.Get(key)
		if value == token || key == token {
			t.Fatal("raw access token was stored in Redis")
		}
	}
	if err := store.Authorize(ctx, token, 77); err != nil {
		t.Fatalf("authorize correct order: %v", err)
	}
	if err := store.Authorize(ctx, token, 78); !errors.Is(err, bookingaccess.ErrInvalidToken) {
		t.Fatalf("expected cross-order denial, got %v", err)
	}
	server.FastForward(bookingaccess.AccessTokenTTL)
	if err := store.Authorize(ctx, token, 77); !errors.Is(err, bookingaccess.ErrInvalidToken) {
		t.Fatalf("expected expired token denial, got %v", err)
	}
}

func TestNewBookingChallengeInvalidatesPreviousCode(t *testing.T) {
	server, client := testRedis(t)
	store := bookingstore.NewChallengeStore(client)
	ctx := context.Background()
	first := bookingaccess.Challenge{
		ID:      "6f745a7f-b0cd-443e-934d-b3aa0d9e542c",
		OrderID: 15, OrderNumber: "AHS-74539", Locator: "AHS-74539", Email: "guest@example.com",
	}
	if err := store.Request(ctx, first, "1234"); err != nil {
		t.Fatalf("request first challenge: %v", err)
	}
	server.FastForward(bookingaccess.CodeRequestInterval)
	second := first
	second.ID = "0c42b9e5-76db-4644-9899-228a26eb357d"
	if err := store.Request(ctx, second, "5678"); err != nil {
		t.Fatalf("request second challenge: %v", err)
	}
	if _, err := store.Verify(ctx, first.ID, "1234"); !errors.Is(err, apierror.ErrCodeExpired) {
		t.Fatalf("previous challenge remained active: %v", err)
	}
	if _, err := store.Verify(ctx, second.ID, "5678"); err != nil {
		t.Fatalf("new challenge is not valid: %v", err)
	}
}

func TestFixedWindowRateLimiter(t *testing.T) {
	server, client := testRedis(t)
	limiter := ratelimit.New(client)
	policy := ratelimit.Policy{Name: "test", Limit: 2, Window: bookingaccess.CodeRequestInterval}
	ctx := context.Background()
	for request := 1; request <= 3; request++ {
		decision, err := limiter.Allow(ctx, policy, "ip:127.0.0.1")
		if err != nil {
			t.Fatalf("request %d: %v", request, err)
		}
		if decision.Allowed != (request <= 2) {
			t.Fatalf("request %d: unexpected decision %+v", request, decision)
		}
	}
	server.FastForward(policy.Window)
	decision, err := limiter.Allow(ctx, policy, "ip:127.0.0.1")
	if err != nil || !decision.Allowed {
		t.Fatalf("window did not reset: decision=%+v err=%v", decision, err)
	}
}

func testRedis(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	t.Helper()
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	return server, client
}
