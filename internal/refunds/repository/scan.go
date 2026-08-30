package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/Halturshik/TicketAgregator-API/internal/refunds"
)

type scanner interface {
	Scan(dest ...any) error
}

func scanOrder(row scanner) (*refunds.OrderData, error) {
	var order refunds.OrderData
	var userID sql.NullInt64
	if err := row.Scan(&order.ID, &userID, &order.GuestToken, &order.Status); err != nil {
		return nil, err
	}
	if userID.Valid {
		value := int(userID.Int64)
		order.UserID = &value
	}
	return &order, nil
}

func scanTicket(row scanner) (*refunds.TicketData, error) {
	var ticket refunds.TicketData
	var policy []byte
	if err := row.Scan(
		&ticket.ID,
		&ticket.TicketNumber,
		&ticket.Status,
		&ticket.SupplierCode,
		&ticket.SupplierOfferID,
		&ticket.FareType,
		&ticket.RefundPolicyVersion,
		&policy,
		&ticket.Price,
		&ticket.BonusSpent,
		&ticket.BonusEarned,
		&ticket.PayableAmount,
		&ticket.DepartureAt,
	); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(policy, &ticket.RefundPolicy); err != nil {
		return nil, fmt.Errorf("decode refund policy: %w", err)
	}
	return &ticket, nil
}
