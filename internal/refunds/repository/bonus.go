package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Halturshik/TicketAgregator-API/internal/bonus"
	"github.com/Halturshik/TicketAgregator-API/internal/refunds"
)

func (t *transaction) ApplyBonus(ctx context.Context, params refunds.BonusParams) (*refunds.BonusResult, error) {
	var balance int
	var debt int
	err := t.tx.QueryRowContext(ctx, `
		SELECT bonus_points, bonus_debt
		FROM users
		WHERE id = $1
		FOR UPDATE
	`, params.UserID).Scan(&balance, &debt)
	if err == sql.ErrNoRows {
		return nil, refunds.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("lock refund bonus balance: %w", err)
	}

	debtRepaid := min(params.RestoreAmount, debt)
	debt -= debtRepaid
	balance += params.RestoreAmount - debtRepaid

	revokedFromBalance := min(params.RevokeAmount, balance)
	balance -= revokedFromBalance
	debtCreated := params.RevokeAmount - revokedFromBalance
	debt += debtCreated

	result, err := t.tx.ExecContext(ctx, `
		UPDATE users
		SET bonus_points = $1, bonus_debt = $2
		WHERE id = $3
	`, balance, debt, params.UserID)
	if err != nil {
		return nil, fmt.Errorf("update refund bonus balance: %w", err)
	}
	if err := requireAffectedRows(result, 1); err != nil {
		return nil, fmt.Errorf("update refund bonus balance: %w", err)
	}

	transactions := []struct {
		typeName string
		amount   int
	}{
		{bonus.TransactionTypeRefundRestore, params.RestoreAmount},
		{bonus.TransactionTypeDebtRepay, debtRepaid},
		{bonus.TransactionTypeRefundRevoke, params.RevokeAmount},
		{bonus.TransactionTypeDebtCreate, debtCreated},
	}
	for _, transaction := range transactions {
		if transaction.amount == 0 {
			continue
		}
		if err := t.recordBonusTransaction(ctx, params, transaction.typeName, transaction.amount); err != nil {
			return nil, err
		}
	}

	if debtCreated > 0 {
		result, err := t.tx.ExecContext(ctx, `
			UPDATE refunds
			SET bonus_debt_created = $1
			WHERE id = $2
		`, debtCreated, params.RefundID)
		if err != nil {
			return nil, fmt.Errorf("update refund bonus debt: %w", err)
		}
		if err := requireAffectedRows(result, 1); err != nil {
			return nil, fmt.Errorf("update refund bonus debt: %w", err)
		}
	}

	return &refunds.BonusResult{Balance: balance, Debt: debt, DebtCreated: debtCreated}, nil
}

func (t *transaction) recordBonusTransaction(ctx context.Context, params refunds.BonusParams, transactionType string, amount int) error {
	result, err := t.tx.ExecContext(ctx, `
		INSERT INTO bonus_transactions (user_id, order_id, refund_id, type, amount)
		VALUES ($1, $2, $3, $4, $5)
	`, params.UserID, params.OrderID, params.RefundID, transactionType, amount)
	if err != nil {
		return fmt.Errorf("record %s bonuses: %w", transactionType, err)
	}
	if err := requireAffectedRows(result, 1); err != nil {
		return fmt.Errorf("record %s bonuses: %w", transactionType, err)
	}
	return nil
}
