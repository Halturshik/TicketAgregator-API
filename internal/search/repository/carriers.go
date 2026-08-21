package repository

import (
	"context"
	"fmt"
)

func (r *Repository) ListCarriers(ctx context.Context) ([]Carrier, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, name, code, transport_type
		FROM carriers
		ORDER BY transport_type, name
	`)
	if err != nil {
		return nil, fmt.Errorf("list carriers: %w", err)
	}
	defer rows.Close()

	result := make([]Carrier, 0)
	for rows.Next() {
		var c Carrier
		if err := rows.Scan(&c.ID, &c.Name, &c.Code, &c.TransportType); err != nil {
			return nil, fmt.Errorf("scan carrier: %w", err)
		}
		result = append(result, c)
	}
	return result, rows.Err()
}
