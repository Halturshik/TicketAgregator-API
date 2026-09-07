package trips

type ScheduledTrip struct {
	Transport     string `json:"transport"`
	Carrier       string `json:"carrier"`
	CarrierCode   string `json:"carrier_code"`
	RouteNumber   string `json:"route_number"`
	FromCity      string `json:"from_city"`
	ToCity        string `json:"to_city"`
	DepartureTime string `json:"departure_time"`
	ArrivalTime   string `json:"arrival_time"`
}

type LookupResult struct {
	TicketNumber string          `json:"ticket_number,omitempty"`
	RouteNumber  string          `json:"route_number,omitempty"`
	Date         string          `json:"date,omitempty"`
	Trips        []ScheduledTrip `json:"trips"`
}
