package orders

import "context"

type Service interface {
	Create(ctx context.Context, userID *int, in CreateOrderInput) (*Order, error)
	List(ctx context.Context, userID int, filter ListFilter) ([]Order, error)
}
