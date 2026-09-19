package repository

import (
	"errors"
	"testing"

	"github.com/lib/pq"
)

func TestIsOrderNumberConflictMatchesOnlyOrderNumberConstraint(t *testing.T) {
	conflict := &pq.Error{Code: uniqueViolationCode, Constraint: orderNumberConstraintName}
	if !isOrderNumberConflict(conflict) {
		t.Fatal("order number conflict was not recognized")
	}
	otherConstraint := &pq.Error{Code: uniqueViolationCode, Constraint: "tickets_ticket_number_key"}
	if isOrderNumberConflict(otherConstraint) {
		t.Fatal("another unique constraint was recognized as order number conflict")
	}
	if isOrderNumberConflict(errors.New("storage error")) {
		t.Fatal("generic error was recognized as order number conflict")
	}
}
