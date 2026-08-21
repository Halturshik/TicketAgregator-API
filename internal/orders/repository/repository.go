package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/Halturshik/TicketAgregator-API/internal/orders"
)

type Repository struct {
	DB *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{DB: db}
}

func (r *Repository) GetUserBonus(ctx context.Context, userID int) (int, error) {
	var bonus int
	err := r.DB.QueryRowContext(ctx, `SELECT bonus_points FROM users WHERE id = $1`, userID).Scan(&bonus)
	return bonus, err
}

func (r *Repository) Create(ctx context.Context, params orders.CreateOrderParams) (*orders.Order, error) {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin order tx: %w", err)
	}
	defer tx.Rollback()

	var order orders.Order
	var nullableUserID sql.NullInt64
	if params.UserID != nil {
		nullableUserID = sql.NullInt64{Int64: int64(*params.UserID), Valid: true}
	}
	err = tx.QueryRowContext(ctx, `
		INSERT INTO orders (user_id, guest_email, guest_payment_token, total_price, bonus_spent, bonus_earned, payable_amount)
		VALUES ($1, $2, NULLIF($3, '')::UUID, $4, $5, $6, $7)
		RETURNING id, status, total_price, bonus_spent, bonus_earned, payable_amount
	`, nullableUserID, params.GuestEmail, params.GuestPaymentToken, params.TotalPrice, params.BonusSpent, params.BonusEarned, params.PayableAmount).
		Scan(&order.ID, &order.Status, &order.TotalPrice, &order.BonusSpent, &order.BonusEarned, &order.PayableAmount)
	if err != nil {
		return nil, fmt.Errorf("insert order: %w", err)
	}
	order.UserID = params.UserID
	order.GuestEmail = params.GuestEmail
	order.GuestPaymentToken = params.GuestPaymentToken

	for _, draft := range params.Tickets {
		ticket, err := insertTicket(ctx, tx, order.ID, draft)
		if err != nil {
			return nil, err
		}
		order.Tickets = append(order.Tickets, *ticket)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit order tx: %w", err)
	}
	return &order, nil
}

func insertTicket(ctx context.Context, tx *sql.Tx, orderID int, draft orders.TicketDraft) (*orders.Ticket, error) {
	passengerData, err := json.Marshal(draft.Passenger)
	if err != nil {
		return nil, fmt.Errorf("marshal passenger snapshot: %w", err)
	}
	documentData, err := json.Marshal(draft.Document)
	if err != nil {
		return nil, fmt.Errorf("marshal document snapshot: %w", err)
	}
	var savedPassengerID sql.NullInt64
	if draft.Passenger.SavedPassengerID != nil {
		savedPassengerID = sql.NullInt64{Int64: int64(*draft.Passenger.SavedPassengerID), Valid: true}
	}

	var ticketID int
	err = tx.QueryRowContext(ctx, `
		INSERT INTO tickets
			(order_id, ticket_number, saved_passenger_id, passenger_snapshot, document_snapshot, transport_type, is_international, price)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`, orderID, draft.TicketNumber, savedPassengerID, passengerData, documentData, draft.Transport, draft.IsInternational, draft.Price).Scan(&ticketID)
	if err != nil {
		return nil, fmt.Errorf("insert ticket: %w", err)
	}

	for _, segment := range draft.Segments {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO ticket_segments
				(ticket_id, segment_order, from_city_id, to_city_id, from_city, to_city, departure_time, arrival_time, carrier_id, carrier, carrier_code, route_number)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,NULLIF($9, 0),$10,$11,$12)
		`, ticketID, segment.Order, segment.FromCityID, segment.ToCityID, segment.FromCity, segment.ToCity, segment.DepartureTime, segment.ArrivalTime, segment.CarrierID, segment.Carrier, segment.CarrierCode, segment.RouteNumber)
		if err != nil {
			return nil, fmt.Errorf("insert ticket segment: %w", err)
		}
	}

	return &orders.Ticket{
		ID: ticketID, TicketNumber: draft.TicketNumber, Transport: draft.Transport,
		IsInternational: draft.IsInternational, Price: draft.Price, Status: "booked",
		Passenger: draft.Passenger, Document: draft.Document, Segments: draft.Segments,
	}, nil
}

func (r *Repository) List(ctx context.Context, userID int, filter orders.ListFilter) ([]orders.Order, error) {
	if filter.Limit <= 0 || filter.Limit > 50 {
		filter.Limit = 10
	}
	query := `
		SELECT DISTINCT o.id, o.status, o.total_price, o.bonus_spent, o.bonus_earned, o.payable_amount
		FROM orders o
		LEFT JOIN tickets t ON t.order_id = o.id
		WHERE o.user_id = $1
	`
	args := []any{userID}
	if filter.Transport != "" {
		query += " AND t.transport_type = $2"
		args = append(args, filter.Transport)
	}
	query += fmt.Sprintf(" ORDER BY o.id DESC LIMIT %d OFFSET %d", filter.Limit, filter.Offset)

	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list orders: %w", err)
	}
	defer rows.Close()

	result := []orders.Order{}
	for rows.Next() {
		var o orders.Order
		id := userID
		o.UserID = &id
		if err := rows.Scan(&o.ID, &o.Status, &o.TotalPrice, &o.BonusSpent, &o.BonusEarned, &o.PayableAmount); err != nil {
			return nil, fmt.Errorf("scan order: %w", err)
		}
		result = append(result, o)
	}
	return result, rows.Err()
}
