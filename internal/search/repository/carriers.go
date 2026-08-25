package repository

import (
	"context"
	"fmt"

	"github.com/Halturshik/TicketAgregator-API/internal/search"
)

func (r *Repository) ListCarriers(ctx context.Context) ([]search.CarrierConfig, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, name, code, transport_type
		FROM carriers
		ORDER BY transport_type, name
	`)
	if err != nil {
		return nil, fmt.Errorf("list carriers: %w", err)
	}
	defer rows.Close()

	result := make([]search.CarrierConfig, 0)
	for rows.Next() {
		var carrier search.CarrierConfig
		if err := rows.Scan(&carrier.ID, &carrier.Name, &carrier.Code, &carrier.TransportType); err != nil {
			return nil, fmt.Errorf("scan carrier: %w", err)
		}
		result = append(result, carrier)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate carriers: %w", err)
	}
	return result, nil
}
