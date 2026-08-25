package search

import "context"

type Service interface {
	Search(ctx context.Context, transport string, in SearchInput, userID *int) (*SearchResult, error)
	GetPage(ctx context.Context, searchID string, offset int, limit int, userID *int) (*SearchResult, error)
	GetTripOption(ctx context.Context, searchID string, tripOptionID string) (*TripOption, error)
	GetOffer(ctx context.Context, searchID string, offerID string) (*Offer, error)
	GetCachedResult(ctx context.Context, searchID string) (*CachedResult, error)
	ListCities(ctx context.Context) (*CitiesResult, error)
}

type Repository interface {
	GetCity(ctx context.Context, id int) (*City, error)
	ListCities(ctx context.Context) ([]City, error)
	ListCitiesPublic(ctx context.Context) ([]CityPublic, error)
	ListCarriers(ctx context.Context) ([]CarrierConfig, error)
}

type Store interface {
	Save(ctx context.Context, result *CachedResult) error
	Get(ctx context.Context, id string) (*CachedResult, error)
}
