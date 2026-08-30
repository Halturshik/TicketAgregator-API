package provider

import (
	"fmt"

	"github.com/Halturshik/TicketAgregator-API/internal/fare"
	"github.com/Halturshik/TicketAgregator-API/internal/supplier"
	"github.com/Halturshik/TicketAgregator-API/internal/transport"
)

type SupplierProfile struct {
	Code             string
	CarrierCodes     map[string]map[string]struct{}
	FareTypes        []string
	PriceModifierMin int
	PriceModifierMax int
}

var supplierProfiles = map[string]SupplierProfile{
	supplier.ProviderAtlas: {
		Code: supplier.ProviderAtlas,
		CarrierCodes: carrierCoverage(
			[]string{"SU", "S7", "DP", "U6", "TK"},
			[]string{"RZD", "FPK"},
			[]string{"FLB", "ETP", "ATL"},
		),
		FareTypes:        allFareTypes(),
		PriceModifierMin: 9_600,
		PriceModifierMax: 10_200,
	},
	supplier.ProviderNexus: {
		Code: supplier.ProviderNexus,
		CarrierCodes: carrierCoverage(
			[]string{"SU", "DP", "TK", "EK", "LH", "AF"},
			[]string{"RZD", "SAP"},
			[]string{"FLB", "ATL", "UNT"},
		),
		FareTypes:        allFareTypes(),
		PriceModifierMin: 9_900,
		PriceModifierMax: 10_600,
	},
	supplier.ProviderVertex: {
		Code: supplier.ProviderVertex,
		CarrierCodes: carrierCoverage(
			[]string{"S7", "U6", "EK", "LH", "AF"},
			[]string{"FPK", "SAP"},
			[]string{"ETP", "ATL", "UNT"},
		),
		FareTypes:        allFareTypes(),
		PriceModifierMin: 9_400,
		PriceModifierMax: 10_400,
	},
}

func allFareTypes() []string {
	return []string{fare.NonRefundable, fare.Standard, fare.Flexible}
}

func Profile(code string) (SupplierProfile, error) {
	profile, ok := supplierProfiles[code]
	if !ok {
		return SupplierProfile{}, fmt.Errorf("unknown supplier profile %q", code)
	}
	return profile, nil
}

func carrierCoverage(avia []string, rail []string, bus []string) map[string]map[string]struct{} {
	return map[string]map[string]struct{}{
		transport.Avia: stringSet(avia),
		transport.Rail: stringSet(rail),
		transport.Bus:  stringSet(bus),
	}
}

func stringSet(values []string) map[string]struct{} {
	result := make(map[string]struct{}, len(values))
	for _, value := range values {
		result[value] = struct{}{}
	}
	return result
}
