package service

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/Halturshik/TicketAgregator-API/internal/bonus"
	"github.com/Halturshik/TicketAgregator-API/internal/fare"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
	"github.com/Halturshik/TicketAgregator-API/internal/search"
	"github.com/Halturshik/TicketAgregator-API/internal/supplier"
)

const (
	nonRefundableShare = 40
	standardShare      = 35
	percentScale       = 100
)

type supplierResult struct {
	providerCode string
	items        []search.TripOption
	err          error
}

func (s *Service) generateTripOptions(
	ctx context.Context,
	transport string,
	in search.SearchInput,
	route searchRoute,
	total int,
	seed int64,
) ([]search.TripOption, error) {
	results := make(chan supplierResult, len(s.providerCodes))
	var workers sync.WaitGroup
	for _, providerCode := range s.providerCodes {
		workers.Add(1)
		go func(code string) {
			defer workers.Done()
			items, err := s.supplier.SearchOffers(ctx, supplier.SearchRequest{
				ProviderCode: code, Transport: transport, Input: in,
				From: route.from, To: route.to, Cities: route.cities,
				Carriers: s.carriers, Count: total, Seed: seed,
			})
			results <- supplierResult{providerCode: code, items: items, err: err}
		}(providerCode)
	}
	go func() {
		workers.Wait()
		close(results)
	}()

	all := make([]search.TripOption, 0, total*len(s.providerCodes))
	var firstErr error
	succeeded := 0
	for result := range results {
		if result.err != nil {
			logger.Warn("Поставщик %s не вернул предложения: %v", result.providerCode, result.err)
			if firstErr == nil {
				firstErr = result.err
			}
			continue
		}
		succeeded++
		all = append(all, result.items...)
	}
	if succeeded == 0 {
		return nil, fmt.Errorf("all suppliers failed: %w", firstErr)
	}

	merged := mergeOffers(all)
	selected := selectBalancedOffers(merged, total)
	sort.Slice(selected, func(i, j int) bool {
		if selected[i].Price == selected[j].Price {
			return selected[i].ID < selected[j].ID
		}
		return selected[i].Price < selected[j].Price
	})
	return selected, nil
}

func mergeOffers(items []search.TripOption) []search.TripOption {
	best := make(map[string]search.TripOption, len(items))
	for _, item := range items {
		applyBonus(&item)
		key := item.ScheduleID + ":" + item.FareType
		current, exists := best[key]
		if !exists || item.Price < current.Price || item.Price == current.Price && item.SupplierCode < current.SupplierCode {
			best[key] = item
		}
	}
	result := make([]search.TripOption, 0, len(best))
	for _, item := range best {
		result = append(result, item)
	}
	return result
}

func selectBalancedOffers(items []search.TripOption, limit int) []search.TripOption {
	if len(items) <= limit {
		return items
	}
	buckets := map[string][]search.TripOption{
		fare.NonRefundable: {},
		fare.Standard:      {},
		fare.Flexible:      {},
	}
	for _, item := range items {
		buckets[item.FareType] = append(buckets[item.FareType], item)
	}
	for code := range buckets {
		sort.Slice(buckets[code], func(i, j int) bool { return buckets[code][i].Price < buckets[code][j].Price })
	}

	quotas := map[string]int{
		fare.NonRefundable: limit * nonRefundableShare / percentScale,
		fare.Standard:      limit * standardShare / percentScale,
	}
	quotas[fare.Flexible] = limit - quotas[fare.NonRefundable] - quotas[fare.Standard]

	selected := make([]search.TripOption, 0, limit)
	leftovers := make([]search.TripOption, 0, len(items))
	for _, code := range []string{fare.NonRefundable, fare.Standard, fare.Flexible} {
		take := min(quotas[code], len(buckets[code]))
		selected = append(selected, buckets[code][:take]...)
		leftovers = append(leftovers, buckets[code][take:]...)
	}
	if len(selected) < limit {
		sort.Slice(leftovers, func(i, j int) bool { return leftovers[i].Price < leftovers[j].Price })
		selected = append(selected, leftovers[:min(limit-len(selected), len(leftovers))]...)
	}
	return selected
}

func applyBonus(item *search.TripOption) {
	item.Outbound.BonusEarn = bonus.Earned(item.Outbound.Price)
	if item.Return != nil {
		item.Return.BonusEarn = bonus.Earned(item.Return.Price)
	}
	item.BonusEarn = bonus.Earned(item.Price)
}
