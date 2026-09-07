package repository

import (
	"errors"

	"github.com/lib/pq"
)

const (
	uniqueViolationCode       = "23505"
	orderNumberConstraintName = "orders_order_number_unique"
)

func isOrderNumberConflict(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) &&
		pqErr.Code == uniqueViolationCode &&
		pqErr.Constraint == orderNumberConstraintName
}
