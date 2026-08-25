package payments

import "context"

type Transaction interface {
	Create(ctx context.Context, record Record) (int, error)
}

type Provider interface {
	Process(ctx context.Context) (bool, error)
}
