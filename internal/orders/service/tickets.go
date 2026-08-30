package service

import (
	"fmt"

	"github.com/Halturshik/TicketAgregator-API/internal/orders"
	"github.com/Halturshik/TicketAgregator-API/internal/orders/number"
	"github.com/Halturshik/TicketAgregator-API/internal/search"
)

const refundPolicyVersion = 1

func buildTickets(trip *selectedTrip, passengers []orders.OrderPassengerDraft) ([]orders.TicketDraft, error) {
	tickets := make([]orders.TicketDraft, 0, len(trip.directions)*len(passengers))
	for _, direction := range trip.directions {
		series, err := number.NewSeries(direction.Transport)
		if err != nil {
			return nil, fmt.Errorf("create ticket number series: %w", err)
		}
		for passengerIndex := range passengers {
			ticketNumber, err := series.Next()
			if err != nil {
				return nil, fmt.Errorf("generate ticket number: %w", err)
			}
			tickets = append(tickets, orders.TicketDraft{
				TicketNumber: ticketNumber, PassengerIndex: passengerIndex,
				SupplierCode: trip.supplierCode, SupplierOfferID: trip.supplierOfferID,
				FareType: trip.fareType, RefundPolicy: trip.refundPolicy,
				RefundPolicyVersion: refundPolicyVersion,
				Transport:           direction.Transport, IsInternational: direction.IsInternational,
				Price: direction.PricePerPassenger, Segments: snapshotSegments(direction.Segments),
			})
		}
	}
	return tickets, nil
}

func snapshotSegments(segments []search.Segment) []orders.TicketSegmentSnapshot {
	result := make([]orders.TicketSegmentSnapshot, len(segments))
	for index, segment := range segments {
		result[index] = orders.TicketSegmentSnapshot{
			Order:      segment.Order,
			FromCityID: segment.FromCityID, ToCityID: segment.ToCityID,
			FromCity: segment.FromCity, ToCity: segment.ToCity,
			DepartureTime: segment.DepartureTime, ArrivalTime: segment.ArrivalTime,
			CarrierID: segment.CarrierID, Carrier: segment.Carrier, CarrierCode: segment.CarrierCode,
			RouteNumber: segment.RouteNumber,
		}
	}
	return result
}
