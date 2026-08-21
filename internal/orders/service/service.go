package service

import (
	"context"
	"fmt"
	"math"
	"net/mail"
	"strings"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/documents"
	"github.com/Halturshik/TicketAgregator-API/internal/orders"
	"github.com/Halturshik/TicketAgregator-API/internal/orders/number"
	"github.com/Halturshik/TicketAgregator-API/internal/orders/repository"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
	"github.com/Halturshik/TicketAgregator-API/internal/search"
	"github.com/google/uuid"
)

const bonusRate = 0.02

type Service struct {
	repo      *repository.Repository
	search    search.Service
	documents documents.Service
}

func NewService(repo *repository.Repository, search search.Service, documents documents.Service) orders.Service {
	return &Service{repo: repo, search: search, documents: documents}
}

func (s *Service) Create(ctx context.Context, userID *int, in orders.CreateOrderInput) (*orders.Order, error) {
	if in.SearchID == "" || len(in.OfferIDs) == 0 || len(in.Passengers) == 0 || in.UseBonus < 0 {
		return nil, apierror.ErrInvalidRequest
	}

	result, err := s.search.GetCachedResult(ctx, in.SearchID)
	if err != nil {
		return nil, err
	}
	if result.Input.Passengers != len(in.Passengers) {
		return nil, apierror.ErrPassengerCountMismatch
	}

	offers, err := selectOffers(result.Items, in.OfferIDs)
	if err != nil {
		return nil, err
	}
	tickets, total, err := s.buildTickets(ctx, in.SearchID, offers, in.Passengers)
	if err != nil {
		return nil, err
	}

	bonusSpent, err := s.bonusSpent(ctx, userID, in.UseBonus, total)
	if err != nil {
		return nil, err
	}
	bonusEarned := 0
	if userID != nil {
		bonusEarned = int(math.Floor(float64(total) * bonusRate))
	}

	guestEmail, guestPaymentToken, err := guestPaymentData(userID, in.GuestEmail)
	if err != nil {
		return nil, err
	}

	order, err := s.repo.Create(ctx, orders.CreateOrderParams{
		UserID:            userID,
		GuestEmail:        guestEmail,
		GuestPaymentToken: guestPaymentToken,
		TotalPrice:        total,
		BonusSpent:        bonusSpent,
		BonusEarned:       bonusEarned,
		PayableAmount:     total - bonusSpent,
		Tickets:           tickets,
	})
	if err != nil {
		logger.Error("Ошибка создания заказа: %v", err)
		return nil, err
	}
	logger.Info("Заказ создан: orderID=%d tickets=%d total=%d payable=%d", order.ID, len(order.Tickets), order.TotalPrice, order.PayableAmount)
	return order, nil
}

func (s *Service) buildTickets(ctx context.Context, searchID string, offers []search.Offer, passengers []orders.PassengerBooking) ([]orders.TicketDraft, int, error) {
	tickets := make([]orders.TicketDraft, 0, len(offers)*len(passengers))
	total := 0
	for _, offer := range offers {
		if offer.PricePerPassenger <= 0 || offer.Price != offer.PricePerPassenger*len(passengers) {
			logger.Error(
				"Некорректная цена предложения в кеше: searchID=%s offerID=%s price=%d pricePerPassenger=%d passengers=%d",
				searchID, offer.ID, offer.Price, offer.PricePerPassenger, len(passengers),
			)
			return nil, 0, fmt.Errorf("invalid cached offer price offerID=%s", offer.ID)
		}
		departure, err := offerDeparture(searchID, offer)
		if err != nil {
			return nil, 0, err
		}
		series, err := number.NewSeries(offer.Transport)
		if err != nil {
			return nil, 0, fmt.Errorf("create ticket number series: %w", err)
		}
		for _, booking := range passengers {
			if err := s.documents.ValidateForBooking(ctx, documents.BookingValidationInput{
				Passenger: documents.BookingPassenger{
					FirstName: booking.Passenger.FirstName, MiddleName: booking.Passenger.MiddleName,
					LastName: booking.Passenger.LastName, BirthDate: booking.Passenger.BirthDate,
					IsRussian: booking.Passenger.IsRussian,
				},
				Document:  documents.BookingDocument{Type: booking.Document.Type, Number: booking.Document.Number},
				Transport: offer.Transport, IsInternational: offer.IsInternational, DepartureTime: departure,
			}); err != nil {
				return nil, 0, err
			}
			ticketNumber, err := series.Next()
			if err != nil {
				return nil, 0, fmt.Errorf("generate ticket number: %w", err)
			}
			tickets = append(tickets, orders.TicketDraft{
				TicketNumber: ticketNumber, Transport: offer.Transport, IsInternational: offer.IsInternational,
				Price: offer.PricePerPassenger, Passenger: booking.Passenger, Document: booking.Document,
				Segments: offer.Segments,
			})
		}
		total += offer.Price
	}
	return tickets, total, nil
}

func offerDeparture(searchID string, offer search.Offer) (time.Time, error) {
	if len(offer.Segments) == 0 {
		logger.Error("Предложение в кеше не содержит сегментов: searchID=%s offerID=%s", searchID, offer.ID)
		return time.Time{}, fmt.Errorf("offer has no segments offerID=%s", offer.ID)
	}
	departure, err := time.Parse(time.RFC3339, offer.Segments[0].DepartureTime)
	if err != nil {
		logger.Error(
			"Некорректное время отправления в кеше: searchID=%s offerID=%s departureTime=%q: %v",
			searchID, offer.ID, offer.Segments[0].DepartureTime, err,
		)
		return time.Time{}, fmt.Errorf("parse offer departure offerID=%s: %w", offer.ID, err)
	}
	return departure, nil
}

func selectOffers(all []search.Offer, ids []string) ([]search.Offer, error) {
	byID := make(map[string]search.Offer, len(all))
	for _, offer := range all {
		byID[offer.ID] = offer
	}
	selected := make([]search.Offer, 0, len(ids))
	seen := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		if _, duplicate := seen[id]; duplicate {
			return nil, apierror.ErrInvalidRequest
		}
		offer, ok := byID[id]
		if !ok {
			return nil, apierror.ErrNotFound
		}
		seen[id] = struct{}{}
		selected = append(selected, offer)
	}
	return selected, nil
}

func (s *Service) bonusSpent(ctx context.Context, userID *int, requested int, total int) (int, error) {
	if userID == nil || requested == 0 {
		return 0, nil
	}
	maxSpend := total / 2
	if requested > maxSpend {
		requested = maxSpend
	}
	balance, err := s.repo.GetUserBonus(ctx, *userID)
	if err != nil {
		logger.Error("Ошибка получения баланса бонусов userID=%d: %v", *userID, err)
		return 0, err
	}
	if requested > balance {
		requested = balance
	}
	return requested, nil
}

func guestPaymentData(userID *int, email string) (string, string, error) {
	if userID != nil {
		return "", "", nil
	}
	email = strings.TrimSpace(email)
	if _, err := mail.ParseAddress(email); err != nil {
		return "", "", apierror.ErrInvalidRequest
	}
	return email, uuid.NewString(), nil
}

func (s *Service) List(ctx context.Context, userID int, filter orders.ListFilter) ([]orders.Order, error) {
	items, err := s.repo.List(ctx, userID, filter)
	if err != nil {
		logger.Error("Ошибка получения истории заказов userID=%d: %v", userID, err)
		return nil, err
	}
	return items, nil
}
