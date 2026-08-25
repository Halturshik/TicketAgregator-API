package payments

import "context"

type ProviderFunc func(ctx context.Context) (bool, error)

func (f ProviderFunc) Process(ctx context.Context) (bool, error) {
	return f(ctx)
}
