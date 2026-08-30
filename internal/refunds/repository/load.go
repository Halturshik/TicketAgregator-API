package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Halturshik/TicketAgregator-API/internal/refunds"
	"github.com/lib/pq"
)

const orderColumns = `
	SELECT id, user_id, COALESCE(guest_payment_token::text, ''), status
	FROM orders
	WHERE id = $1
`

const ticketColumns = `
	SELECT t.id, t.ticket_number, t.status, t.supplier_code, t.supplier_offer_id, t.fare_type,
		t.refund_policy_version, t.refund_policy_snapshot, t.price,
		t.bonus_spent, t.bonus_earned, t.payable_amount,
		(SELECT MIN(ts.departure_time) FROM ticket_segments ts WHERE ts.ticket_id = t.id)
	FROM tickets t
`

func (r *Repository) Load(ctx context.Context, orderID int, ticketIDs []int, all bool) (*refunds.OrderData, []refunds.TicketData, error) {
	order, err := scanOrder(r.db.QueryRowContext(ctx, orderColumns, orderID))
	if err == sql.ErrNoRows {
		return nil, nil, refunds.ErrNotFound
	}
	if err != nil {
		return nil, nil, fmt.Errorf("load refund order: %w", err)
	}

	tickets, err := loadTickets(ctx, r.db, orderID, ticketIDs, all, false)
	if err != nil {
		return nil, nil, err
	}
	return order, tickets, nil
}

type queryer interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

func loadTickets(ctx context.Context, db queryer, orderID int, ticketIDs []int, all bool, lock bool) ([]refunds.TicketData, error) {
	query := ticketColumns + " WHERE t.order_id = $1"
	args := []any{orderID}
	if all {
		query += " AND t.status = 'paid'"
	} else {
		query += " AND t.id = ANY($2)"
		args = append(args, pq.Array(ticketIDs))
	}
	query += " ORDER BY t.id"
	if lock {
		query += " FOR UPDATE OF t"
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("load refund tickets: %w", err)
	}
	defer rows.Close()

	tickets := make([]refunds.TicketData, 0, len(ticketIDs))
	for rows.Next() {
		ticket, err := scanTicket(rows)
		if err != nil {
			return nil, fmt.Errorf("scan refund ticket: %w", err)
		}
		tickets = append(tickets, *ticket)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate refund tickets: %w", err)
	}
	return tickets, nil
}

func (t *transaction) LockOrder(ctx context.Context, orderID int) (*refunds.OrderData, error) {
	order, err := scanOrder(t.tx.QueryRowContext(ctx, orderColumns+" FOR UPDATE", orderID))
	if err == sql.ErrNoRows {
		return nil, refunds.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("lock refund order: %w", err)
	}
	return order, nil
}

func (t *transaction) LockTickets(ctx context.Context, orderID int, ticketIDs []int, all bool) ([]refunds.TicketData, error) {
	return loadTickets(ctx, t.tx, orderID, ticketIDs, all, true)
}
