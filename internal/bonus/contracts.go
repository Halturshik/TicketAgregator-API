package bonus

import "context"

type Service interface {
	List(ctx context.Context, userID int, limit int, offset int) ([]Transaction, error)
}

type Repository interface {
	List(ctx context.Context, userID int, limit int, offset int) ([]Transaction, error)
}

type BalanceReader interface {
	GetBalance(ctx context.Context, userID int) (int, error)
}

type PaymentTransaction interface {
	Spend(ctx context.Context, userID int, orderID int, amount int) error
	Earn(ctx context.Context, userID int, orderID int, amount int) error
	GetBalance(ctx context.Context, userID int) (int, error)
}
