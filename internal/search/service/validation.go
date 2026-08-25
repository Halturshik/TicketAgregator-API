package service

import (
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/search"
)

func normalizeInput(in *search.SearchInput, now time.Time) (time.Time, error) {
	if in.Passengers < MinPassengers || in.Passengers > MaxPassengers {
		return time.Time{}, apierror.ErrInvalidRequest
	}
	if in.Limit <= 0 || in.Limit > MaxPageSize {
		in.Limit = DefaultPageSize
	}
	searchDate, err := time.Parse(search.DateLayout, in.Date)
	if err != nil {
		return time.Time{}, apierror.ErrInvalidRequest
	}
	today := dateOnly(now)
	maxDate := today.AddDate(0, MaxSearchMonths, 0)
	if searchDate.Before(today) || searchDate.After(maxDate) {
		return time.Time{}, apierror.ErrInvalidRequest
	}
	if in.ReturnDate != "" {
		returnDate, err := time.Parse(search.DateLayout, in.ReturnDate)
		if err != nil || !returnDate.After(searchDate) || returnDate.After(maxDate) {
			return time.Time{}, apierror.ErrInvalidRequest
		}
	}
	return searchDate, nil
}

func dateOnly(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}
