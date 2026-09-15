package requestid

import (
	"context"

	"github.com/google/uuid"
)

const (
	Header      = "X-Request-ID"
	MetadataKey = "x-request-id"
)

type contextKey struct{}

func New() string {
	return uuid.NewString()
}

func Resolve(value string) string {
	id, err := uuid.Parse(value)
	if err != nil {
		return New()
	}
	return id.String()
}

func WithContext(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, contextKey{}, id)
}

func FromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(contextKey{}).(string)
	return id, ok && id != ""
}
