package search

type SearchResult struct {
	SearchID string       `json:"search_id"`
	Total    int          `json:"total"`
	Offset   int          `json:"offset"`
	Limit    int          `json:"limit"`
	Items    []TripOption `json:"items"`
}

type CachedResult struct {
	SearchID string       `json:"search_id"`
	Input    SearchInput  `json:"input"`
	Total    int          `json:"total"`
	Items    []TripOption `json:"items"`
}
