package payments

import "context"

type Service interface {
	Pay(ctx context.Context, userID *int, orderID int, guestPaymentToken string) (*PayOutput, error)
}
