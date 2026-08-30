package provider

import "github.com/Halturshik/TicketAgregator-API/internal/search"

type Request struct {
	Input  search.SearchInput
	From   search.City
	To     search.City
	Cities []search.City
	Count  int
	Seed   int64
}
