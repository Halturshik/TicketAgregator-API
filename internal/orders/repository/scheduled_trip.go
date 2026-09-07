package repository

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/Halturshik/TicketAgregator-API/internal/orders"
)

func insertScheduledTrip(
	ctx context.Context,
	tx *sql.Tx,
	tripKey string,
	transport string,
	segment orders.TicketSegmentSnapshot,
) (int, error) {
	var id int
	err := tx.QueryRowContext(ctx, `
		INSERT INTO scheduled_trips
			(trip_key, transport_type, from_city_id, to_city_id, from_city, to_city,
			 departure_time, arrival_time, carrier_id, carrier, carrier_code, route_number)
		VALUES ($1, $2, NULLIF($3, 0), NULLIF($4, 0), $5, $6, $7, $8,
			NULLIF($9, 0), $10, $11, $12)
		ON CONFLICT (trip_key) DO UPDATE SET trip_key = EXCLUDED.trip_key
		RETURNING id
	`, tripKey, transport, segment.FromCityID, segment.ToCityID,
		segment.FromCity, segment.ToCity, segment.DepartureTime, segment.ArrivalTime,
		segment.CarrierID, segment.Carrier, segment.CarrierCode, segment.RouteNumber).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert scheduled trip: %w", err)
	}
	return id, nil
}

func scheduledTripKey(transport string, segment orders.TicketSegmentSnapshot) string {
	parts := []string{
		strings.ToLower(strings.TrimSpace(transport)),
		strings.ToUpper(strings.TrimSpace(segment.CarrierCode)),
		strings.ToUpper(strings.TrimSpace(segment.RouteNumber)),
		strconv.Itoa(segment.FromCityID),
		strconv.Itoa(segment.ToCityID),
		strings.TrimSpace(segment.DepartureTime),
	}
	sum := sha256.Sum256([]byte(strings.Join(parts, "\x00")))
	return hex.EncodeToString(sum[:])
}
