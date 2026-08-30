package provider

import (
	"context"
	"hash/fnv"
	"math/rand"

	"github.com/Halturshik/TicketAgregator-API/internal/bonus"
	"github.com/Halturshik/TicketAgregator-API/internal/fare"
	"github.com/Halturshik/TicketAgregator-API/internal/search"
	"github.com/google/uuid"
)

const priceBasisPoints = 10_000

var farePriceModifiers = map[string]int{
	fare.NonRefundable: 9_000,
	fare.Standard:      10_000,
	fare.Flexible:      11_500,
}

func GenerateSupplierOffers(
	ctx context.Context,
	providerCode string,
	transport string,
	carriers []Carrier,
	req Request,
) ([]search.TripOption, error) {
	profile, err := Profile(providerCode)
	if err != nil {
		return nil, err
	}
	catalog, err := NewMockProvider(transport, carriers).Generate(ctx, req)
	if err != nil {
		return nil, err
	}

	result := make([]search.TripOption, 0, len(catalog))
	for _, option := range catalog {
		if !profile.supportsOption(option) {
			continue
		}
		result = append(result, applySupplierProfile(option, profile, req.Input.Passengers))
	}
	return result, nil
}

func (p SupplierProfile) supportsOption(option search.TripOption) bool {
	allowed := p.CarrierCodes[option.Transport]
	if len(option.Outbound.Segments) == 0 {
		return false
	}
	_, ok := allowed[option.Outbound.Segments[0].CarrierCode]
	return ok
}

func applySupplierProfile(option search.TripOption, profile SupplierProfile, passengers int) search.TripOption {
	rnd := rand.New(rand.NewSource(profileSeed(option.ScheduleID, profile.Code)))
	fareType := profile.FareTypes[rnd.Intn(len(profile.FareTypes))]
	policy, _ := fare.Policy(fareType)
	providerModifier := profile.PriceModifierMin
	if extra := profile.PriceModifierMax - profile.PriceModifierMin; extra > 0 {
		providerModifier += rnd.Intn(extra + 1)
	}
	fareModifier := farePriceModifiers[fareType]

	applyOfferPrice(&option.Outbound, passengers, fareModifier, providerModifier)
	option.Price = option.Outbound.Price
	option.PricePerPassenger = option.Outbound.PricePerPassenger
	if option.Return != nil {
		applyOfferPrice(option.Return, passengers, fareModifier, providerModifier)
		option.Price += option.Return.Price
		option.PricePerPassenger += option.Return.PricePerPassenger
	}

	offerID := uuid.NewSHA1(
		uuid.NameSpaceOID,
		[]byte(option.ScheduleID+":"+profile.Code+":"+fareType),
	).String()
	option.ID = offerID
	option.SupplierCode = profile.Code
	option.SupplierOfferID = offerID
	option.FareType = fareType
	option.RefundPolicy = policy
	option.BonusEarn = bonus.Earned(option.Price)
	return option
}

func applyOfferPrice(offer *search.Offer, passengers int, fareModifier int, providerModifier int) {
	price := offer.PricePerPassenger
	price = max(1, price*fareModifier/priceBasisPoints)
	price = max(1, price*providerModifier/priceBasisPoints)
	offer.PricePerPassenger = price
	offer.Price = price * passengers
	offer.BonusEarn = bonus.Earned(offer.Price)
}

func profileSeed(scheduleID string, providerCode string) int64 {
	hash := fnv.New64a()
	_, _ = hash.Write([]byte(scheduleID))
	_, _ = hash.Write([]byte(providerCode))
	return int64(hash.Sum64())
}
