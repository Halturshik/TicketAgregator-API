package search

import "context"

type Service interface {
	Search(ctx context.Context, transport string, in SearchInput, userID *int) (*SearchResult, error)
	GetPage(ctx context.Context, searchID string, offset int, limit int, userID *int) (*SearchResult, error)
	GetOffer(ctx context.Context, searchID string, offerID string) (*Offer, error)
	GetCachedResult(ctx context.Context, searchID string) (*CachedResult, error)
	ListCities(ctx context.Context) (*CitiesResult, error)
}
