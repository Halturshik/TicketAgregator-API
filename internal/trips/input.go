package trips

type LookupInput struct {
	TicketNumber string `json:"ticket_number,omitempty"`
	RouteNumber  string `json:"route_number,omitempty"`
	Date         string `json:"date,omitempty"`
}
