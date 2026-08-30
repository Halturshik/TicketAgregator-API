package search

import "github.com/Halturshik/TicketAgregator-API/internal/fare"

type Segment struct {
	Order         int    `json:"order"`
	FromCityID    int    `json:"from_city_id"`
	ToCityID      int    `json:"to_city_id"`
	FromCity      string `json:"from_city"`
	ToCity        string `json:"to_city"`
	DepartureTime string `json:"departure_time"`
	ArrivalTime   string `json:"arrival_time"`
	CarrierID     int    `json:"carrier_id"`
	Carrier       string `json:"carrier"`
	CarrierCode   string `json:"carrier_code"`
	RouteNumber   string `json:"route_number"`
}

type Offer struct {
	ID                string    `json:"id"`
	Transport         string    `json:"transport"`
	IsInternational   bool      `json:"is_international"`
	Price             int       `json:"price"`
	PricePerPassenger int       `json:"price_per_passenger"`
	BonusEarn         int       `json:"bonus_earn"`
	BonusHint         string    `json:"bonus_hint,omitempty"`
	TransferCount     int       `json:"transfer_count"`
	DurationMinutes   int       `json:"duration_minutes"`
	Segments          []Segment `json:"segments"`
}

type TripOption struct {
	ID                string            `json:"id"`
	ScheduleID        string            `json:"schedule_id"`
	SupplierCode      string            `json:"supplier_code"`
	SupplierOfferID   string            `json:"supplier_offer_id"`
	FareType          string            `json:"fare_type"`
	RefundPolicy      fare.RefundPolicy `json:"refund_policy"`
	Transport         string            `json:"transport"`
	Price             int               `json:"price"`
	PricePerPassenger int               `json:"price_per_passenger"`
	BonusEarn         int               `json:"bonus_earn"`
	BonusHint         string            `json:"bonus_hint,omitempty"`
	Outbound          Offer             `json:"outbound"`
	Return            *Offer            `json:"return,omitempty"`
}
