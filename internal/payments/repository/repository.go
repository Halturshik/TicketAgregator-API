package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

var (
	ErrOrderCannotBePaid = errors.New("order cannot be paid")
	ErrPaymentForbidden  = errors.New("payment forbidden")
	ErrInsufficientBonus = errors.New("insufficient bonus")
)

type Repository struct {
	DB *sql.DB
}

type OrderPaymentData struct {
	ID            int
	UserID        sql.NullInt64
	GuestToken    sql.NullString
	Status        string
	PayableAmount int
	BonusSpent    int
	BonusEarned   int
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{DB: db}
}

func (r *Repository) Pay(ctx context.Context, orderID int, expectedUserID *int, guestPaymentToken string, success bool) (paymentID int, data OrderPaymentData, balance *int, err error) {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return 0, data, nil, fmt.Errorf("begin payment tx: %w", err)
	}
	defer tx.Rollback()

	err = tx.QueryRowContext(ctx, `
		SELECT id, user_id, guest_payment_token, status, payable_amount, bonus_spent, bonus_earned
		FROM orders
		WHERE id = $1
		FOR UPDATE
	`, orderID).Scan(&data.ID, &data.UserID, &data.GuestToken, &data.Status, &data.PayableAmount, &data.BonusSpent, &data.BonusEarned)
	if err != nil {
		return 0, data, nil, err
	}
	if data.UserID.Valid {
		if expectedUserID == nil || int(data.UserID.Int64) != *expectedUserID {
			return 0, data, nil, ErrPaymentForbidden
		}
	} else if !data.GuestToken.Valid || guestPaymentToken == "" || data.GuestToken.String != guestPaymentToken {
		return 0, data, nil, ErrPaymentForbidden
	}
	if data.Status != "created" {
		return 0, data, nil, ErrOrderCannotBePaid
	}

	status := "failed"
	if success {
		status = "success"
	}
	err = tx.QueryRowContext(ctx, `
		INSERT INTO payments (order_id, amount, status, provider)
		VALUES ($1, $2, $3, 'mock')
		RETURNING id
	`, orderID, data.PayableAmount, status).Scan(&paymentID)
	if err != nil {
		return 0, data, nil, fmt.Errorf("insert payment: %w", err)
	}

	if success {
		_, err = tx.ExecContext(ctx, `UPDATE orders SET status = 'paid', paid_at = NOW() WHERE id = $1`, orderID)
		if err != nil {
			return 0, data, nil, fmt.Errorf("mark order paid: %w", err)
		}
		_, err = tx.ExecContext(ctx, `UPDATE tickets SET status = 'paid' WHERE order_id = $1`, orderID)
		if err != nil {
			return 0, data, nil, fmt.Errorf("mark tickets paid: %w", err)
		}
		if data.UserID.Valid {
			uid := int(data.UserID.Int64)
			if data.BonusSpent > 0 {
				var updatedBalance int
				err = tx.QueryRowContext(ctx, `
					UPDATE users
					SET bonus_points = bonus_points - $1
					WHERE id = $2 AND bonus_points >= $1
					RETURNING bonus_points
				`, data.BonusSpent, uid).Scan(&updatedBalance)
				if err == sql.ErrNoRows {
					return 0, data, nil, ErrInsufficientBonus
				}
				if err != nil {
					return 0, data, nil, fmt.Errorf("spend bonuses: %w", err)
				}
				_, _ = tx.ExecContext(ctx, `INSERT INTO bonus_transactions (user_id, order_id, type, amount) VALUES ($1,$2,'spend',$3)`, uid, orderID, data.BonusSpent)
			}
			if data.BonusEarned > 0 {
				_, err = tx.ExecContext(ctx, `UPDATE users SET bonus_points = bonus_points + $1 WHERE id = $2`, data.BonusEarned, uid)
				if err != nil {
					return 0, data, nil, fmt.Errorf("earn bonuses: %w", err)
				}
				_, _ = tx.ExecContext(ctx, `INSERT INTO bonus_transactions (user_id, order_id, type, amount) VALUES ($1,$2,'earn',$3)`, uid, orderID, data.BonusEarned)
			}
			var b int
			if err := tx.QueryRowContext(ctx, `SELECT bonus_points FROM users WHERE id = $1`, uid).Scan(&b); err == nil {
				balance = &b
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, data, nil, fmt.Errorf("commit payment tx: %w", err)
	}
	return paymentID, data, balance, nil
}
