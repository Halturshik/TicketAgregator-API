package store

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestRequestCodeIsRateLimitedAtomically(t *testing.T) {
	_, client := newTestRedis(t)
	store := NewCodeService(client)

	const workers = 32
	start := make(chan struct{})
	results := make(chan error, workers)
	var group sync.WaitGroup
	group.Add(workers)
	for range workers {
		go func() {
			defer group.Done()
			<-start
			results <- store.RequestCode(context.Background(), "user@example.com", "123456")
		}()
	}
	close(start)
	group.Wait()
	close(results)

	succeeded := 0
	rateLimited := 0
	for err := range results {
		switch {
		case err == nil:
			succeeded++
		case errors.Is(err, apierror.ErrCodeRateLimited):
			rateLimited++
		default:
			t.Fatalf("unexpected RequestCode() error = %v", err)
		}
	}
	if succeeded != 1 || rateLimited != workers-1 {
		t.Fatalf("results = success:%d limited:%d, want 1 and %d", succeeded, rateLimited, workers-1)
	}
}

func TestVerifyCodeAppliesCooldownAfterMaximumAttempts(t *testing.T) {
	mini, client := newTestRedis(t)
	store := NewCodeService(client)
	ctx := context.Background()
	email := "attempts@example.com"
	if err := store.RequestCode(ctx, email, "123456"); err != nil {
		t.Fatalf("RequestCode() error = %v", err)
	}

	for attempt := 1; attempt < MaxVerifyAttempts; attempt++ {
		if err := store.Verify(ctx, email, "000000"); !errors.Is(err, apierror.ErrInvalidVerificationCode) {
			t.Fatalf("attempt %d error = %v, want invalid code", attempt, err)
		}
	}
	if err := store.Verify(ctx, email, "000000"); !errors.Is(err, apierror.ErrTooManyAttempts) {
		t.Fatalf("last attempt error = %v, want too many attempts", err)
	}
	if err := store.Verify(ctx, email, "123456"); !errors.Is(err, apierror.ErrTooManyAttempts) {
		t.Fatalf("correct code during cooldown error = %v, want too many attempts", err)
	}

	mini.FastForward(CooldownAfterFailed + time.Second)
	if err := store.RequestCode(ctx, email, "654321"); err != nil {
		t.Fatalf("RequestCode() after cooldown error = %v", err)
	}
	if err := store.Verify(ctx, email, "654321"); err != nil {
		t.Fatalf("Verify() after cooldown error = %v", err)
	}
}

func newTestRedis(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	t.Helper()
	mini := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	return mini, client
}
