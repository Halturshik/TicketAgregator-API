package supplier

import "github.com/Halturshik/TicketAgregator-API/internal/search"

type SearchRequest struct {
	ProviderCode string
	Transport    string
	Input        search.SearchInput
	From         search.City
	To           search.City
	Cities       []search.City
	Carriers     []search.CarrierConfig
	Count        int
	Seed         int64
}

type TripOption = search.TripOption
