package search

type SearchInput struct {
	FromCityID int    `json:"from_city_id"`
	ToCityID   int    `json:"to_city_id"`
	Date       string `json:"date"`
	ReturnDate string `json:"return_date,omitempty"`
	Passengers int    `json:"passengers"`
	Offset     int    `json:"offset,omitempty"`
	Limit      int    `json:"limit,omitempty"`
}
