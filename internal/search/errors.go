package search

import "errors"

var (
	ErrCityNotFound                 = errors.New("city not found")
	ErrCachedResultNotFound         = errors.New("cached search result not found")
	ErrInvalidSupplierConfiguration = errors.New("invalid supplier configuration")
)
