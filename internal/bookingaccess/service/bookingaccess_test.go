package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/bookingaccess"
	bookingservice "github.com/Halturshik/TicketAgregator-API/internal/bookingaccess/service"
	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/orders"
	"github.com/Halturshik/TicketAgregator-API/internal/refunds"
)

func TestBookingAccessPublicLookupMasksPersonalData(t *testing.T) {
	fixture := bookingFixture()
	service := newBookingService(&bookingRepositoryStub{booking: fixture})

	ticket, err := service.Lookup(context.Background(), bookingaccess.LookupInput{TicketNumber: "AV-20260821"})
	if err != nil {
		t.Fatalf("ticket lookup: %v", err)
	}
	if ticket.Kind != "ticket" || ticket.OrderNumber != fixture.OrderNumber {
		t.Fatalf("unexpected public ticket summary: %+v", ticket)
	}
	if ticket.Ticket == nil || ticket.Ticket.Passenger != "Иван П." {
		t.Fatalf("passenger was not masked: %+v", ticket.Ticket)
	}
	if len(ticket.Ticket.Segments) != 2 || ticket.Directions[0].TransferCount != 1 {
		t.Fatalf("transfer itinerary was lost: %+v", ticket)
	}

	order, err := service.Lookup(context.Background(), bookingaccess.LookupInput{OrderNumber: fixture.OrderNumber})
	if err != nil {
		t.Fatalf("order lookup: %v", err)
	}
	if order.Ticket != nil || order.TicketCount != 2 || order.PassengerCount != 2 {
		t.Fatalf("unexpected public order summary: %+v", order)
	}
	if len(order.Directions) != 2 {
		t.Fatalf("expected outbound and return directions, got %d", len(order.Directions))
	}
}

func TestBookingAccessRequestDoesNotRevealEmailMismatch(t *testing.T) {
	repo := &bookingRepositoryStub{booking: bookingFixture(), guestNotFound: true}
	mailer := &mailerStub{}
	service := newBookingServiceWith(repo, &challengeStoreStub{}, &accessStoreStub{}, mailer, &refundServiceStub{})

	result, err := service.RequestAccess(context.Background(), bookingaccess.AccessRequestInput{
		OrderNumber: repo.booking.OrderNumber,
		Email:       "wrong@example.com",
	})
	if err != nil {
		t.Fatalf("request access: %v", err)
	}
	if result.ChallengeID == "" || result.Message == "" {
		t.Fatalf("generic response is incomplete: %+v", result)
	}
	if mailer.calls != 0 {
		t.Fatal("email must not be sent for a mismatched locator")
	}
	_, err = service.ConfirmAccess(context.Background(), bookingaccess.AccessConfirmInput{
		ChallengeID: result.ChallengeID, Code: "1234",
	})
	if !errors.Is(err, apierror.ErrCodeExpired) {
		t.Fatalf("dummy challenge must never grant access, got %v", err)
	}
}

func TestBookingAccessChallengeDetailsAndRefundFacade(t *testing.T) {
	fixture := bookingFixture()
	challengeStore := &challengeStoreStub{}
	accessStore := &accessStoreStub{token: "guest-access-token"}
	mailer := &mailerStub{}
	refundService := &refundServiceStub{}
	service := newBookingServiceWith(
		&bookingRepositoryStub{booking: fixture}, challengeStore, accessStore, mailer, refundService,
	)

	requested, err := service.RequestAccess(context.Background(), bookingaccess.AccessRequestInput{
		TicketNumber: fixture.Tickets[0].Number,
		Email:        "GUEST@EXAMPLE.COM",
	})
	if err != nil {
		t.Fatalf("request access: %v", err)
	}
	if mailer.to != fixture.GuestEmail || mailer.code != "1234" {
		t.Fatalf("unexpected email: %+v", mailer)
	}
	if challengeStore.challenge.OrderID != fixture.ID {
		t.Fatalf("challenge is not scoped to order: %+v", challengeStore.challenge)
	}

	confirmed, err := service.ConfirmAccess(context.Background(), bookingaccess.AccessConfirmInput{
		ChallengeID: requested.ChallengeID,
		Code:        "1234",
	})
	if err != nil {
		t.Fatalf("confirm access: %v", err)
	}
	if confirmed.Token != accessStore.token || confirmed.OrderNumber != fixture.OrderNumber {
		t.Fatalf("unexpected access result: %+v", confirmed)
	}

	details, err := service.Details(context.Background(), nil, fixture.OrderNumber, accessStore.token)
	if err != nil {
		t.Fatalf("booking details: %v", err)
	}
	if details.Tickets[0].Passenger != "Иван П." || details.Tickets[0].Document != "**** 4567" {
		t.Fatalf("expanded details leaked or lost masked data: %+v", details.Tickets[0])
	}

	quote, err := service.QuoteRefund(context.Background(), nil, fixture.OrderNumber, accessStore.token,
		bookingaccess.RefundInput{TicketNumbers: []string{fixture.Tickets[1].Number}})
	if err != nil {
		t.Fatalf("quote refund: %v", err)
	}
	if len(refundService.lastInput.TicketIDs) != 1 || refundService.lastInput.TicketIDs[0] != fixture.Tickets[1].ID {
		t.Fatalf("public ticket number was mapped incorrectly: %+v", refundService.lastInput)
	}
	if refundService.lastInput.GuestPaymentToken != fixture.GuestPaymentToken {
		t.Fatal("original guest payment token was not supplied internally")
	}
	if quote.Items[0].TicketNumber != fixture.Tickets[1].Number {
		t.Fatalf("numeric ticket id leaked into quote: %+v", quote.Items[0])
	}

	result, err := service.Refund(context.Background(), nil, fixture.OrderNumber, accessStore.token,
		bookingaccess.RefundInput{All: true}, "81fb6f4d-d0d9-45e1-a94d-4d7843445a22")
	if err != nil {
		t.Fatalf("refund: %v", err)
	}
	if !refundService.lastInput.All || refundService.lastInput.IdempotencyKey == "" {
		t.Fatalf("refund facade lost operation parameters: %+v", refundService.lastInput)
	}
	if result.OrderNumber != fixture.OrderNumber || result.Items[0].TicketNumber == "" {
		t.Fatalf("unexpected public refund result: %+v", result)
	}
}

func TestBookingAccessRejectsCrossOrderAndHidesForeignOrder(t *testing.T) {
	fixture := bookingFixture()
	access := &accessStoreStub{authorizeErr: bookingaccess.ErrInvalidToken}
	service := newBookingServiceWith(
		&bookingRepositoryStub{booking: fixture}, &challengeStoreStub{}, access,
		&mailerStub{}, &refundServiceStub{},
	)
	_, err := service.Details(context.Background(), nil, fixture.OrderNumber, "token-for-another-order")
	if !errors.Is(err, apierror.ErrInvalidToken) {
		t.Fatalf("expected invalid token, got %v", err)
	}

	owner := 11
	fixture.UserID = &owner
	wrongUser := 12
	_, err = service.Details(context.Background(), &wrongUser, fixture.OrderNumber, "")
	if !errors.Is(err, apierror.ErrNotFound) {
		t.Fatalf("expected hidden foreign order, got %v", err)
	}
}

func bookingFixture() *bookingaccess.Booking {
	departure := time.Date(2026, 10, 10, 8, 0, 0, 0, time.UTC)
	return &bookingaccess.Booking{
		ID: 7, OrderNumber: "AHS-74539", GuestEmail: "guest@example.com",
		GuestPaymentToken: "b59d8297-8893-4f70-9779-dabf150627b8",
		Status:            orders.OrderStatusPaid, TotalPrice: 20000, CurrentTotalPrice: 20000,
		BonusSpent: 1000, BonusEarned: 400, CreatedAt: departure.Add(-24 * time.Hour),
		PassengerCount: 2,
		Tickets: []bookingaccess.Ticket{
			{
				ID: 101, Number: "AV-20260821", Status: orders.TicketStatusPaid,
				Transport: "avia", FareType: "standard", Price: 10000,
				Passenger: orders.PassengerSnapshot{FirstName: "Иван", LastName: "Петров"},
				Document:  orders.DocumentSnapshot{Type: "international_passport", Number: "721234567"},
				Segments: []bookingaccess.Segment{
					{Order: 1, Carrier: "Aeroflot", CarrierCode: "SU", RouteNumber: "SU 12345", FromCity: "Москва", ToCity: "Стамбул", DepartureTime: departure, ArrivalTime: departure.Add(3 * time.Hour)},
					{Order: 2, Carrier: "Turkish Airlines", CarrierCode: "TK", RouteNumber: "TK 54321", FromCity: "Стамбул", ToCity: "Париж", DepartureTime: departure.Add(5 * time.Hour), ArrivalTime: departure.Add(9 * time.Hour)},
				},
			},
			{
				ID: 102, Number: "AB-20269999", Status: orders.TicketStatusPaid,
				Transport: "avia", FareType: "flexible", Price: 10000,
				Passenger: orders.PassengerSnapshot{FirstName: "Анна", LastName: "Смирнова"},
				Document:  orders.DocumentSnapshot{Type: "international_passport", Number: "729876543"},
				Segments: []bookingaccess.Segment{
					{Order: 1, Carrier: "Aeroflot", CarrierCode: "SU", RouteNumber: "SU 99999", FromCity: "Париж", ToCity: "Москва", DepartureTime: departure.AddDate(0, 0, 7), ArrivalTime: departure.AddDate(0, 0, 7).Add(4 * time.Hour)},
				},
			},
		},
	}
}

func newBookingService(repo bookingaccess.Repository) bookingaccess.Service {
	return newBookingServiceWith(repo, &challengeStoreStub{}, &accessStoreStub{}, &mailerStub{}, &refundServiceStub{})
}

func newBookingServiceWith(
	repo bookingaccess.Repository,
	challenges bookingaccess.ChallengeStore,
	access bookingaccess.AccessStore,
	mailer bookingaccess.Mailer,
	refundService bookingaccess.RefundService,
) bookingaccess.Service {
	return bookingservice.NewService(repo, challenges, access, mailer, codeGeneratorStub{}, refundService)
}

type bookingRepositoryStub struct {
	booking       *bookingaccess.Booking
	guestNotFound bool
}

func (s *bookingRepositoryStub) PublicByTicket(context.Context, string) (*bookingaccess.Booking, error) {
	return s.booking, nil
}
func (s *bookingRepositoryStub) PublicByOrder(context.Context, string) (*bookingaccess.Booking, error) {
	return s.booking, nil
}
func (s *bookingRepositoryStub) GuestByLocator(context.Context, bookingaccess.Locator, string) (*bookingaccess.Booking, error) {
	if s.guestNotFound {
		return nil, bookingaccess.ErrNotFound
	}
	return s.booking, nil
}
func (s *bookingRepositoryStub) ByOrderNumber(context.Context, string) (*bookingaccess.Booking, error) {
	return s.booking, nil
}

type challengeStoreStub struct {
	challenge bookingaccess.Challenge
	code      string
}

func (s *challengeStoreStub) Request(_ context.Context, challenge bookingaccess.Challenge, code string) error {
	s.challenge, s.code = challenge, code
	return nil
}
func (s *challengeStoreStub) Verify(_ context.Context, id string, code string) (*bookingaccess.Challenge, error) {
	if s.challenge.ID != id || s.code != code {
		return nil, apierror.ErrInvalidVerificationCode
	}
	return &s.challenge, nil
}

type accessStoreStub struct {
	token        string
	authorizeErr error
}

func (s *accessStoreStub) Issue(context.Context, int) (string, error) {
	if s.token == "" {
		s.token = "issued-token"
	}
	return s.token, nil
}
func (s *accessStoreStub) Authorize(_ context.Context, token string, _ int) error {
	if s.authorizeErr != nil {
		return s.authorizeErr
	}
	if token != s.token {
		return bookingaccess.ErrInvalidToken
	}
	return nil
}

type mailerStub struct {
	calls int
	to    string
	code  string
}

func (s *mailerStub) SendVerificationEmail(_ context.Context, to string, code string) error {
	s.calls++
	s.to, s.code = to, code
	return nil
}

type codeGeneratorStub struct{}

func (codeGeneratorStub) GenerateVerificationCode() (string, error) { return "1234", nil }

type refundServiceStub struct {
	lastInput refunds.Input
}

func (s *refundServiceStub) Quote(_ context.Context, _ *int, orderID int, input refunds.Input) (*refunds.QuoteOutput, error) {
	s.lastInput = input
	ticketID := 102
	if len(input.TicketIDs) > 0 {
		ticketID = input.TicketIDs[0]
	}
	return &refunds.QuoteOutput{
		OrderID: orderID, Refundable: true, CashAmount: 7000,
		Items: []refunds.QuoteItem{{TicketID: ticketID, Refundable: true, RefundPercent: 70, GrossAmount: 10000, GrossRefundAmount: 7000}},
	}, nil
}
func (s *refundServiceStub) Refund(_ context.Context, _ *int, orderID int, input refunds.Input) (*refunds.Result, error) {
	s.lastInput = input
	return &refunds.Result{
		OrderID: orderID, OrderStatus: orders.OrderStatusRefunded, Status: refunds.StatusProcessing,
		Items: []refunds.QuoteItem{{TicketID: 101, Refundable: true}},
	}, nil
}
