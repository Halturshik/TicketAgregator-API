package bonus

import "context"

type Service interface {
	List(ctx context.Context, userID int, limit int, offset int) ([]Transaction, error)
}
