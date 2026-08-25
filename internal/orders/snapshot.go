package orders

type PassengerSnapshot struct {
	FirstName  string `json:"first_name"`
	MiddleName string `json:"middle_name,omitempty"`
	LastName   string `json:"last_name"`
	BirthDate  string `json:"birth_date"`
	IsRussian  bool   `json:"is_russian"`
}

type DocumentSnapshot struct {
	Type               string `json:"type"`
	Number             string `json:"number"`
	ExpiresAt          string `json:"expires_at,omitempty"`
	VerificationStatus string `json:"verification_status"`
	LastCheckedAt      string `json:"last_checked_at"`
}

type TicketSegmentSnapshot struct {
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
