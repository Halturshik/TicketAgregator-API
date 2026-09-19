package store

import (
	"context"
	"sync"
	"testing"
)

func TestResetVerificationCanBeConsumedOnce(t *testing.T) {
	_, client := newTestRedis(t)
	store := NewResetStore(client)
	ctx := context.Background()
	email := "reset@example.com"
	if err := store.SaveVerified(ctx, email); err != nil {
		t.Fatalf("SaveVerified() error = %v", err)
	}

	const workers = 16
	start := make(chan struct{})
	results := make(chan bool, workers)
	errors := make(chan error, workers)
	var group sync.WaitGroup
	group.Add(workers)
	for range workers {
		go func() {
			defer group.Done()
			<-start
			consumed, err := store.ConsumeVerified(ctx, email)
			results <- consumed
			errors <- err
		}()
	}
	close(start)
	group.Wait()
	close(results)
	close(errors)

	for err := range errors {
		if err != nil {
			t.Fatalf("ConsumeVerified() error = %v", err)
		}
	}
	succeeded := 0
	for consumed := range results {
		if consumed {
			succeeded++
		}
	}
	if succeeded != 1 {
		t.Fatalf("successful consumptions = %d, want 1", succeeded)
	}
	if verified, err := store.IsVerified(ctx, email); err != nil || verified {
		t.Fatalf("IsVerified() = (%t, %v), want (false, nil)", verified, err)
	}
}
