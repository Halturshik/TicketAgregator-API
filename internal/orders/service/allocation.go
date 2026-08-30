package service

import (
	"fmt"
	"sort"

	"github.com/Halturshik/TicketAgregator-API/internal/orders"
)

type allocationPart struct {
	index     int
	remainder int
}

func allocateTicketPricing(tickets []orders.TicketDraft, pricing *orderPricing, total int) error {
	if total <= 0 || len(tickets) == 0 {
		return fmt.Errorf("invalid ticket pricing input")
	}
	spent := allocateByPrice(tickets, pricing.spent, total)
	earned := allocateByPrice(tickets, pricing.earned, total)
	for index := range tickets {
		if spent[index] > tickets[index].Price {
			return fmt.Errorf("allocated bonus exceeds ticket price")
		}
		tickets[index].BonusSpent = spent[index]
		tickets[index].BonusEarned = earned[index]
		tickets[index].PayableAmount = tickets[index].Price - spent[index]
	}
	return nil
}

func allocateByPrice(tickets []orders.TicketDraft, amount int, total int) []int {
	result := make([]int, len(tickets))
	parts := make([]allocationPart, len(tickets))
	allocated := 0
	for index, ticket := range tickets {
		value := amount * ticket.Price
		result[index] = value / total
		allocated += result[index]
		parts[index] = allocationPart{index: index, remainder: value % total}
	}
	sort.SliceStable(parts, func(i, j int) bool { return parts[i].remainder > parts[j].remainder })
	for index := 0; index < amount-allocated; index++ {
		result[parts[index%len(parts)].index]++
	}
	return result
}
