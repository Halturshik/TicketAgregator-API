package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Halturshik/TicketAgregator-API/internal/bonus"
)

func (t *transaction) Spend(ctx context.Context, userID int, orderID int, amount int) error {
	var updatedBalance int
	err := t.tx.QueryRowContext(ctx, `
		UPDATE users
		SET bonus_points = bonus_points - $1
		WHERE id = $2 AND bonus_points >= $1
		RETURNING bonus_points
	`, amount, userID).Scan(&updatedBalance)
	if err == sql.ErrNoRows {
		return bonus.ErrInsufficientBalance
	}
	if err != nil {
		return fmt.Errorf("spend bonuses: %w", err)
	}
	return t.recordBonusTransaction(ctx, userID, orderID, bonus.TransactionTypeSpend, amount)
}

func (t *transaction) Earn(ctx context.Context, userID int, orderID int, amount int) error {
	var balance int
	var debt int
	err := t.tx.QueryRowContext(ctx, `
		SELECT bonus_points, bonus_debt
		FROM users
		WHERE id = $1
		FOR UPDATE
	`, userID).Scan(&balance, &debt)
	if err == sql.ErrNoRows {
		return fmt.Errorf("earn bonuses: user not found")
	}
	if err != nil {
		return fmt.Errorf("lock bonus debt before earning: %w", err)
	}

	debtRepaid := min(amount, debt)
	debt -= debtRepaid
	balance += amount - debtRepaid
	result, err := t.tx.ExecContext(ctx, `
		UPDATE users
		SET bonus_points = $1, bonus_debt = $2
		WHERE id = $3
	`, balance, debt, userID)
	if err != nil {
		return fmt.Errorf("earn bonuses: %w", err)
	}
	if err := requireAffectedRows(result, 1); err != nil {
		return fmt.Errorf("earn bonuses: %w", err)
	}
	if err := t.recordBonusTransaction(ctx, userID, orderID, bonus.TransactionTypeEarn, amount); err != nil {
		return err
	}
	if debtRepaid > 0 {
		return t.recordBonusTransaction(ctx, userID, orderID, bonus.TransactionTypeDebtRepay, debtRepaid)
	}
	return nil
}

func (t *transaction) recordBonusTransaction(ctx context.Context, userID int, orderID int, transactionType string, amount int) error {
	result, err := t.tx.ExecContext(ctx, `
		INSERT INTO bonus_transactions (user_id, order_id, type, amount)
		VALUES ($1, $2, $3, $4)
	`, userID, orderID, transactionType, amount)
	if err != nil {
		return fmt.Errorf("record %s bonuses: %w", transactionType, err)
	}
	if err := requireAffectedRows(result, 1); err != nil {
		return fmt.Errorf("record %s bonuses: %w", transactionType, err)
	}
	return nil
}

func (t *transaction) GetBalance(ctx context.Context, userID int) (int, error) {
	var balance int
	if err := t.tx.QueryRowContext(ctx, `SELECT bonus_points FROM users WHERE id = $1`, userID).Scan(&balance); err != nil {
		return 0, fmt.Errorf("get bonus balance after payment: %w", err)
	}
	return balance, nil
}
