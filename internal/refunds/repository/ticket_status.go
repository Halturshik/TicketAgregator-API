package repository

import (
	"context"
	"fmt"

	"github.com/Halturshik/TicketAgregator-API/internal/orders"
	"github.com/lib/pq"
)

func (t *transaction) MarkTicketsPending(ctx context.Context, ticketIDs []int) error {
	return t.updateTicketStatus(ctx, ticketIDs, orders.TicketStatusPaid, orders.TicketStatusRefundPending)
}

func (t *transaction) MarkTicketsRefunded(ctx context.Context, ticketIDs []int) error {
	return t.updateTicketStatus(ctx, ticketIDs, orders.TicketStatusRefundPending, orders.TicketStatusRefunded)
}

func (t *transaction) RestoreTicketsPaid(ctx context.Context, ticketIDs []int) error {
	return t.updateTicketStatus(ctx, ticketIDs, orders.TicketStatusRefundPending, orders.TicketStatusPaid)
}

func (t *transaction) updateTicketStatus(ctx context.Context, ticketIDs []int, from, to string) error {
	result, err := t.tx.ExecContext(ctx, `
		UPDATE tickets
		SET status = $1
		WHERE id = ANY($2) AND status = $3
	`, to, pq.Array(ticketIDs), from)
	if err != nil {
		return fmt.Errorf("update refund ticket status %s -> %s: %w", from, to, err)
	}
	if err := requireAffectedRows(result, int64(len(ticketIDs))); err != nil {
		return fmt.Errorf("update refund ticket status %s -> %s: %w", from, to, err)
	}
	return nil
}
