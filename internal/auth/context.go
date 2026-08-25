package auth

import "context"

type userIDContextKey struct{}

func WithUserID(ctx context.Context, userID int) context.Context {
	return context.WithValue(ctx, userIDContextKey{}, userID)
}

func UserIDFromContext(ctx context.Context) (int, bool) {
	userID, ok := ctx.Value(userIDContextKey{}).(int)
	return userID, ok
}
