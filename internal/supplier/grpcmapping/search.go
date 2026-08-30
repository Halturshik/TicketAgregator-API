package grpcmapping

import (
	supplierv1 "github.com/Halturshik/TicketAgregator-API/internal/gen/supplier/v1"
	"github.com/Halturshik/TicketAgregator-API/internal/search"
	"github.com/Halturshik/TicketAgregator-API/internal/supplier"
)

func SearchRequestToProto(request supplier.SearchRequest) *supplierv1.SearchOffersRequest {
	cities := make([]*supplierv1.City, 0, len(request.Cities))
	for _, city := range request.Cities {
		cities = append(cities, cityToProto(city))
	}
	carriers := make([]*supplierv1.Carrier, 0, len(request.Carriers))
	for _, carrier := range request.Carriers {
		carriers = append(carriers, &supplierv1.Carrier{
			Id: int32(carrier.ID), Name: carrier.Name, Code: carrier.Code, Transport: carrier.TransportType,
		})
	}
	return &supplierv1.SearchOffersRequest{
		ProviderCode: request.ProviderCode,
		Transport:    request.Transport,
		Input: &supplierv1.SearchInput{
			FromId: int32(request.Input.FromCityID), ToId: int32(request.Input.ToCityID),
			Date: request.Input.Date, ReturnDate: request.Input.ReturnDate, Passengers: int32(request.Input.Passengers),
		},
		From: cityToProto(request.From), To: cityToProto(request.To), Cities: cities,
		Carriers: carriers, Count: int32(request.Count), Seed: request.Seed,
	}
}

func SearchRequestFromProto(request *supplierv1.SearchOffersRequest) supplier.SearchRequest {
	input := search.SearchInput{}
	if request.Input != nil {
		input = search.SearchInput{
			FromCityID: int(request.Input.FromId), ToCityID: int(request.Input.ToId),
			Date: request.Input.Date, ReturnDate: request.Input.ReturnDate, Passengers: int(request.Input.Passengers),
		}
	}
	cities := make([]search.City, 0, len(request.Cities))
	for _, city := range request.Cities {
		cities = append(cities, cityFromProto(city))
	}
	carriers := make([]search.CarrierConfig, 0, len(request.Carriers))
	for _, carrier := range request.Carriers {
		carriers = append(carriers, search.CarrierConfig{
			ID: int(carrier.Id), Name: carrier.Name, Code: carrier.Code, TransportType: carrier.Transport,
		})
	}
	return supplier.SearchRequest{
		ProviderCode: request.ProviderCode, Transport: request.Transport, Input: input,
		From: cityFromProto(request.From), To: cityFromProto(request.To), Cities: cities,
		Carriers: carriers, Count: int(request.Count), Seed: request.Seed,
	}
}

func TripOptionsToProto(items []search.TripOption) []*supplierv1.TripOption {
	result := make([]*supplierv1.TripOption, 0, len(items))
	for _, item := range items {
		result = append(result, tripOptionToProto(item))
	}
	return result
}

func TripOptionsFromProto(items []*supplierv1.TripOption) []search.TripOption {
	result := make([]search.TripOption, 0, len(items))
	for _, item := range items {
		result = append(result, tripOptionFromProto(item))
	}
	return result
}

func tripOptionToProto(item search.TripOption) *supplierv1.TripOption {
	result := &supplierv1.TripOption{
		Id: item.ID, ScheduleId: item.ScheduleID, SupplierCode: item.SupplierCode,
		SupplierOfferId: item.SupplierOfferID, FareType: item.FareType,
		RefundPolicy: PolicyToProto(item.RefundPolicy), Transport: item.Transport,
		Price: int32(item.Price), PricePerPassenger: int32(item.PricePerPassenger),
		Outbound: offerToProto(item.Outbound),
	}
	if item.Return != nil {
		result.Return = offerToProto(*item.Return)
	}
	return result
}

func tripOptionFromProto(item *supplierv1.TripOption) search.TripOption {
	result := search.TripOption{
		ID: item.Id, ScheduleID: item.ScheduleId, SupplierCode: item.SupplierCode,
		SupplierOfferID: item.SupplierOfferId, FareType: item.FareType,
		RefundPolicy: PolicyFromProto(item.RefundPolicy), Transport: item.Transport,
		Price: int(item.Price), PricePerPassenger: int(item.PricePerPassenger),
		Outbound: offerFromProto(item.Outbound),
	}
	if item.Return != nil {
		value := offerFromProto(item.Return)
		result.Return = &value
	}
	return result
}

func offerToProto(offer search.Offer) *supplierv1.Offer {
	segments := make([]*supplierv1.Segment, 0, len(offer.Segments))
	for _, segment := range offer.Segments {
		segments = append(segments, &supplierv1.Segment{
			Order: int32(segment.Order), FromCityId: int32(segment.FromCityID), ToCityId: int32(segment.ToCityID),
			FromCity: segment.FromCity, ToCity: segment.ToCity,
			DepartureTime: segment.DepartureTime, ArrivalTime: segment.ArrivalTime,
			CarrierId: int32(segment.CarrierID), Carrier: segment.Carrier,
			CarrierCode: segment.CarrierCode, RouteNumber: segment.RouteNumber,
		})
	}
	return &supplierv1.Offer{
		Id: offer.ID, Transport: offer.Transport, IsInternational: offer.IsInternational,
		Price: int32(offer.Price), PricePerPassenger: int32(offer.PricePerPassenger),
		TransferCount: int32(offer.TransferCount), DurationMinutes: int32(offer.DurationMinutes), Segments: segments,
	}
}

func offerFromProto(offer *supplierv1.Offer) search.Offer {
	if offer == nil {
		return search.Offer{}
	}
	segments := make([]search.Segment, 0, len(offer.Segments))
	for _, segment := range offer.Segments {
		segments = append(segments, search.Segment{
			Order: int(segment.Order), FromCityID: int(segment.FromCityId), ToCityID: int(segment.ToCityId),
			FromCity: segment.FromCity, ToCity: segment.ToCity,
			DepartureTime: segment.DepartureTime, ArrivalTime: segment.ArrivalTime,
			CarrierID: int(segment.CarrierId), Carrier: segment.Carrier,
			CarrierCode: segment.CarrierCode, RouteNumber: segment.RouteNumber,
		})
	}
	return search.Offer{
		ID: offer.Id, Transport: offer.Transport, IsInternational: offer.IsInternational,
		Price: int(offer.Price), PricePerPassenger: int(offer.PricePerPassenger),
		TransferCount: int(offer.TransferCount), DurationMinutes: int(offer.DurationMinutes), Segments: segments,
	}
}

func cityToProto(city search.City) *supplierv1.City {
	neighbors := make([]int32, len(city.NeighborCountryIDs))
	for index, id := range city.NeighborCountryIDs {
		neighbors[index] = int32(id)
	}
	return &supplierv1.City{
		Id: int32(city.ID), Name: city.Name, CountryId: int32(city.CountryID), Country: city.Country,
		IsRussia: city.IsRussia, IsAirHub: city.IsAirHub, Latitude: city.Latitude, Longitude: city.Longitude,
		NeighborCountryIds: neighbors,
	}
}

func cityFromProto(city *supplierv1.City) search.City {
	if city == nil {
		return search.City{}
	}
	neighbors := make([]int, len(city.NeighborCountryIds))
	for index, id := range city.NeighborCountryIds {
		neighbors[index] = int(id)
	}
	return search.City{
		ID: int(city.Id), Name: city.Name, CountryID: int(city.CountryId), Country: city.Country,
		IsRussia: city.IsRussia, IsAirHub: city.IsAirHub, Latitude: city.Latitude, Longitude: city.Longitude,
		NeighborCountryIDs: neighbors,
	}
}
