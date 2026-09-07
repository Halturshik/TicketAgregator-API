package repository

import (
	"fmt"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/orders"
)

func updateHistoryDirection(direction *orders.HistoryDirection, row historyTicketRow) {
	if row.segmentOrder == 1 || direction.DepartureTime == "" {
		direction.FromCity = row.fromCity
		direction.DepartureTime = row.departureTime.UTC().Format(time.RFC3339)
	}
	direction.ToCity = row.toCity
	direction.ArrivalTime = row.arrivalTime.UTC().Format(time.RFC3339)
}

func buildHistoryDirections(items []orders.HistoryOrder) {
	for orderIndex := range items {
		seen := make(map[string]struct{})
		for _, ticket := range items[orderIndex].Tickets {
			direction := ticket.Direction
			key := fmt.Sprintf(
				"%s\x00%s\x00%s\x00%s",
				direction.FromCity, direction.ToCity, direction.DepartureTime, direction.ArrivalTime,
			)
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = struct{}{}
			items[orderIndex].Directions = append(items[orderIndex].Directions, direction)
		}
	}
}
