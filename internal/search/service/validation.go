package service

import (
	"strings"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/search"
	"github.com/Halturshik/TicketAgregator-API/internal/transport"
)

func validProviderCodes(codes []string) bool {
	if len(codes) == 0 {
		return false
	}
	seen := make(map[string]struct{}, len(codes))
	for _, code := range codes {
		code = strings.TrimSpace(code)
		if code == "" {
			return false
		}
		if _, duplicate := seen[code]; duplicate {
			return false
		}
		seen[code] = struct{}{}
	}
	return true
}

func validCarrierConfiguration(carriers []search.CarrierConfig) bool {
	configured := make(map[string]bool, 3)
	for _, carrier := range carriers {
		if strings.TrimSpace(carrier.Name) == "" || strings.TrimSpace(carrier.Code) == "" || !transport.IsSupported(carrier.TransportType) {
			return false
		}
		configured[carrier.TransportType] = true
	}
	return configured[transport.Avia] && configured[transport.Rail] && configured[transport.Bus]
}

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
