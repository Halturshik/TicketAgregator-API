package store

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/Halturshik/TicketAgregator-API/internal/auth/token"
	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
)

func TestRefreshRotationConsumesTokenExactlyOnce(t *testing.T) {
	_, client := newTestRedis(t)
	store := NewRefreshStore(client)
	ctx := context.Background()
	const userID int64 = 42
	const oldToken = "old-refresh-token"
	if err := store.Save(ctx, userID, oldToken); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if err := store.AddToUserSet(ctx, userID, token.HashToken(oldToken)); err != nil {
		t.Fatalf("AddToUserSet() error = %v", err)
	}

	const workers = 16
	start := make(chan struct{})
	results := make(chan error, workers)
	var group sync.WaitGroup
	group.Add(workers)
	for index := range workers {
		go func() {
			defer group.Done()
			<-start
			results <- store.Rotate(ctx, userID, oldToken, fmt.Sprintf("new-refresh-token-%d", index))
		}()
	}
	close(start)
	group.Wait()
	close(results)

	succeeded := 0
	rejected := 0
	for err := range results {
		switch {
		case err == nil:
			succeeded++
		case errors.Is(err, apierror.ErrInvalidToken):
			rejected++
		default:
			t.Fatalf("unexpected Rotate() error = %v", err)
		}
	}
	if succeeded != 1 || rejected != workers-1 {
		t.Fatalf("rotation results = success:%d rejected:%d, want 1 and %d", succeeded, rejected, workers-1)
	}

	if client.Exists(ctx, RefreshKey(token.HashToken(oldToken))).Val() != 0 {
		t.Fatal("consumed refresh token still exists")
	}
	members, err := client.SMembers(ctx, RefreshUserSetKey(userID)).Result()
	if err != nil {
		t.Fatalf("SMembers() error = %v", err)
	}
	if len(members) != 1 {
		t.Fatalf("active refresh set size = %d, want 1", len(members))
	}
	if client.Exists(ctx, RefreshKey(members[0])).Val() != 1 {
		t.Fatal("winning refresh token is absent")
	}
}

func TestRefreshRotationRejectsWrongUserWithoutConsumingToken(t *testing.T) {
	_, client := newTestRedis(t)
	store := NewRefreshStore(client)
	ctx := context.Background()
	const oldToken = "old-refresh-token"
	if err := store.Save(ctx, 42, oldToken); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	if err := store.Rotate(ctx, 99, oldToken, "new-refresh-token"); !errors.Is(err, apierror.ErrInvalidToken) {
		t.Fatalf("Rotate() error = %v, want invalid token", err)
	}
	if client.Exists(ctx, RefreshKey(token.HashToken(oldToken))).Val() != 1 {
		t.Fatal("token was consumed by rotation for another user")
	}
}
