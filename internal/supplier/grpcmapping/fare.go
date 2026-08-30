package grpcmapping

import (
	"github.com/Halturshik/TicketAgregator-API/internal/fare"
	supplierv1 "github.com/Halturshik/TicketAgregator-API/internal/gen/supplier/v1"
)

func PolicyToProto(policy fare.RefundPolicy) *supplierv1.RefundPolicy {
	tiers := make([]*supplierv1.RefundTier, 0, len(policy.Tiers))
	for _, tier := range policy.Tiers {
		tiers = append(tiers, &supplierv1.RefundTier{
			MinHoursBeforeDeparture: int32(tier.MinHoursBeforeDeparture),
			RefundPercent:           int32(tier.RefundPercent),
		})
	}
	return &supplierv1.RefundPolicy{Code: policy.Code, Refundable: policy.Refundable, Tiers: tiers}
}

func PolicyFromProto(policy *supplierv1.RefundPolicy) fare.RefundPolicy {
	if policy == nil {
		return fare.RefundPolicy{}
	}
	tiers := make([]fare.RefundTier, 0, len(policy.Tiers))
	for _, tier := range policy.Tiers {
		tiers = append(tiers, fare.RefundTier{
			MinHoursBeforeDeparture: int(tier.MinHoursBeforeDeparture),
			RefundPercent:           int(tier.RefundPercent),
		})
	}
	return fare.RefundPolicy{Code: policy.Code, Refundable: policy.Refundable, Tiers: tiers}
}
