package search

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

type CityPublic struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Country string `json:"country"`
	IsHub   bool   `json:"is_air_hub"`
}

type CitiesResult struct {
	Items []CityPublic `json:"items"`
}
