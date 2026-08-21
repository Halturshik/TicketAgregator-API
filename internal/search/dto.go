package search

type CityPublic struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Country string `json:"country"`
	IsHub   bool   `json:"is_air_hub"`
}

type CitiesResult struct {
	Items []CityPublic `json:"items"`
}

type SearchInput struct {
	FromCityID int    `json:"from_city_id"`
	ToCityID   int    `json:"to_city_id"`
	Date       string `json:"date"`
	Passengers int    `json:"passengers"`
	Offset     int    `json:"offset,omitempty"`
	Limit      int    `json:"limit,omitempty"`
}

type City struct {
	ID                 int     `json:"id"`
	Name               string  `json:"name"`
	CountryID          int     `json:"country_id"`
	Country            string  `json:"country"`
	IsRussia           bool    `json:"is_russia"`
	IsAirHub           bool    `json:"is_air_hub"`
	Latitude           float64 `json:"-"`
	Longitude          float64 `json:"-"`
	NeighborCountryIDs []int   `json:"-"`
}

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

type SearchResult struct {
	SearchID string  `json:"search_id"`
	Total    int     `json:"total"`
	Offset   int     `json:"offset"`
	Limit    int     `json:"limit"`
	Items    []Offer `json:"items"`
}

type CachedResult struct {
	SearchID string      `json:"search_id"`
	Input    SearchInput `json:"input"`
	Total    int         `json:"total"`
	Items    []Offer     `json:"items"`
}
