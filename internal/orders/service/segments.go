package service

import (
	"fmt"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/search"
)

func validateSegments(offer search.Offer, usedRouteNumbers map[string]struct{}) error {
	var previousArrival time.Time
	var previousToCityID int
	for index, segment := range offer.Segments {
		if segment.Order != index+1 || segment.RouteNumber == "" {
			return fmt.Errorf("invalid segment order or route number")
		}
		if _, duplicate := usedRouteNumbers[segment.RouteNumber]; duplicate {
			return fmt.Errorf("duplicate route number %s", segment.RouteNumber)
		}
		usedRouteNumbers[segment.RouteNumber] = struct{}{}

		departure, err := time.Parse(time.RFC3339, segment.DepartureTime)
		if err != nil {
			return fmt.Errorf("parse departure: %w", err)
		}
		arrival, err := time.Parse(time.RFC3339, segment.ArrivalTime)
		if err != nil || !arrival.After(departure) {
			return fmt.Errorf("invalid arrival time")
		}
		if index > 0 && (segment.FromCityID != previousToCityID || departure.Before(previousArrival)) {
			return fmt.Errorf("disconnected or overlapping transfer")
		}
		previousToCityID = segment.ToCityID
		previousArrival = arrival
	}
	return nil
}

func offerDeparture(offer search.Offer) (time.Time, error) {
	departure, err := time.Parse(time.RFC3339, offer.Segments[0].DepartureTime)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse offer departure offerID=%s: %w", offer.ID, err)
	}
	return departure, nil
}
