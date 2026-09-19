package auth

import (
	"context"
	"testing"
)

func TestUserIDContext(t *testing.T) {
	if _, ok := UserIDFromContext(context.Background()); ok {
		t.Fatal("empty context unexpectedly contains user ID")
	}

	ctx := WithUserID(context.Background(), 42)
	userID, ok := UserIDFromContext(ctx)
	if !ok || userID != 42 {
		t.Fatalf("UserIDFromContext() = (%d, %t), want (42, true)", userID, ok)
	}
}
