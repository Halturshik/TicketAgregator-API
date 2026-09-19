package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/documents"
	"github.com/Halturshik/TicketAgregator-API/internal/fare"
	"github.com/Halturshik/TicketAgregator-API/internal/orders"
	"github.com/Halturshik/TicketAgregator-API/internal/passengers"
	"github.com/Halturshik/TicketAgregator-API/internal/search"
)

func TestCreateRoundTripBuildsTicketPerDirectionAndPassenger(t *testing.T) {
	now := time.Date(2026, time.August, 22, 12, 0, 0, 0, time.UTC)
	result := cachedRoundTrip(2)
	repo := &fakeOrderRepository{}
	documents := &fakeDocumentValidator{}
	svc := &Service{
		repo: repo, search: &fakeSearchReader{result: result}, documents: documents,
		passengers: &fakePassengerReader{}, bonus: &fakeBonusReader{balance: 100},
		now: func() time.Time { return now },
	}
	userID := 7
	in := orders.CreateOrderInput{
		SearchID: "search-1", TripOptionID: "trip-1", UseBonus: 80,
		Passengers: []orders.PassengerBooking{
			newPassenger("Иван", "Иванов", "1234567890"),
			newPassenger("Петр", "Петров", "1234567891"),
		},
	}

	_, err := svc.Create(context.Background(), &userID, in)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	params := repo.created
	if len(params.Passengers) != 2 {
		t.Fatalf("saved order passenger count = %d, want 2", len(params.Passengers))
	}
	if len(params.Tickets) != 4 {
		t.Fatalf("ticket count = %d, want 4", len(params.Tickets))
	}
	if documents.calls != 4 {
		t.Fatalf("document validation calls = %d, want 4", documents.calls)
	}
	if params.TotalPrice != 6000 || params.CurrentTotalPrice != 6000 || params.PayableAmount != 5920 {
		t.Fatalf("order amounts = total:%d current:%d payable:%d",
			params.TotalPrice, params.CurrentTotalPrice, params.PayableAmount)
	}
	if params.BonusSpent != 80 || params.BonusEarned != 120 {
		t.Fatalf("bonuses = spent:%d earned:%d", params.BonusSpent, params.BonusEarned)
	}
	if matched, _ := regexp.MatchString(`^[A-Z]{3}-[0-9]{5}$`, params.OrderNumber); !matched {
		t.Fatalf("order number = %q", params.OrderNumber)
	}
	if !params.ExpiresAt.Equal(now.Add(orders.OrderTTL)) {
		t.Fatalf("expires at = %s, want %s", params.ExpiresAt, now.Add(orders.OrderTTL))
	}

	seenTicketNumbers := make(map[string]struct{})
	ticketPrice := 0
	for _, ticket := range params.Tickets {
		if _, exists := seenTicketNumbers[ticket.TicketNumber]; exists {
			t.Fatalf("duplicate ticket number %s", ticket.TicketNumber)
		}
		seenTicketNumbers[ticket.TicketNumber] = struct{}{}
		ticketPrice += ticket.Price
	}
	if ticketPrice != params.TotalPrice {
		t.Fatalf("tickets price = %d, order total = %d", ticketPrice, params.TotalPrice)
	}
	if params.Tickets[0].Segments[0].RouteNumber != params.Tickets[1].Segments[0].RouteNumber {
		t.Fatal("passengers on outbound received different routes")
	}
	if params.Tickets[2].Segments[0].RouteNumber != params.Tickets[3].Segments[0].RouteNumber {
		t.Fatal("passengers on return received different routes")
	}
	if params.Tickets[0].Segments[0].RouteNumber == params.Tickets[2].Segments[0].RouteNumber {
		t.Fatal("outbound and return unexpectedly use the same route number")
	}
}

func TestCreateSavedPassengerChecksOwnershipAndUsesStoredDataWhenSnapshotEmpty(t *testing.T) {
	result := cachedRoundTrip(1)
	repo := &fakeOrderRepository{}
	documents := &fakeDocumentValidator{}
	passengerReader := &fakePassengerReader{passenger: &passengers.Passenger{
		ID: 9, FirstName: "Иван", LastName: "Иванов", BirthDate: "1990-01-01", IsRussian: true,
	}}
	svc := &Service{repo: repo, search: &fakeSearchReader{result: result}, documents: documents, passengers: passengerReader, now: time.Now}
	userID := 5
	passengerID := 9
	in := orders.CreateOrderInput{
		SearchID: "search-1", TripOptionID: "trip-1",
		Passengers: []orders.PassengerBooking{{
			Source: passengers.SourceSaved, SavedPassengerID: &passengerID,
			Document: orders.DocumentBooking{Type: "internal_passport", Number: "1234567890"},
		}},
	}

	_, err := svc.Create(context.Background(), &userID, in)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if passengerReader.ownerID != userID || passengerReader.passengerID != passengerID {
		t.Fatalf("ownership lookup = owner:%d passenger:%d", passengerReader.ownerID, passengerReader.passengerID)
	}
	if got := repo.created.Passengers[0].Passenger.FirstName; got != "Иван" {
		t.Fatalf("stored snapshot first name = %q, want Иван", got)
	}
	if documents.lastInput.SavedPassengerID == nil || *documents.lastInput.SavedPassengerID != passengerID {
		t.Fatal("document ownership was not scoped to saved passenger")
	}
}

func TestCreateRejectsSavedPassengerForGuest(t *testing.T) {
	passengerID := 1
	svc := testOrderService(cachedRoundTrip(1))
	in := orders.CreateOrderInput{
		SearchID: "search-1", TripOptionID: "trip-1", GuestEmail: "guest@example.com",
		Passengers: []orders.PassengerBooking{{
			Source: passengers.SourceSaved, SavedPassengerID: &passengerID,
		}},
	}
	_, err := svc.Create(context.Background(), nil, in)
	if !errors.Is(err, apierror.ErrUnauthorized) {
		t.Fatalf("Create() error = %v, want unauthorized", err)
	}
}

func TestCreateRejectsDuplicateSavedPassenger(t *testing.T) {
	result := cachedRoundTrip(2)
	passengerID := 9
	reader := &fakePassengerReader{passenger: &passengers.Passenger{
		ID: 9, FirstName: "Иван", LastName: "Иванов", BirthDate: "1990-01-01", IsRussian: true,
	}}
	svc := &Service{
		repo: &fakeOrderRepository{}, search: &fakeSearchReader{result: result},
		documents: &fakeDocumentValidator{}, passengers: reader, bonus: &fakeBonusReader{}, now: time.Now,
	}
	userID := 1
	booking := orders.PassengerBooking{
		Source: passengers.SourceSaved, SavedPassengerID: &passengerID,
		Document: orders.DocumentBooking{Type: "internal_passport", Number: "1234567890"},
	}
	_, err := svc.Create(context.Background(), &userID, orders.CreateOrderInput{
		SearchID: "search-1", TripOptionID: "trip-1", Passengers: []orders.PassengerBooking{booking, booking},
	})
	if !errors.Is(err, apierror.ErrInvalidRequest) {
		t.Fatalf("Create() error = %v, want invalid request", err)
	}
}

func TestCreateRejectsPassengerCountMismatch(t *testing.T) {
	svc := testOrderService(cachedRoundTrip(2))
	_, err := svc.Create(context.Background(), nil, orders.CreateOrderInput{
		SearchID: "search-1", TripOptionID: "trip-1", GuestEmail: "guest@example.com",
		Passengers: []orders.PassengerBooking{newPassenger("Ivan", "Ivanov", "1234567890")},
	})
	if !errors.Is(err, apierror.ErrPassengerCountMismatch) {
		t.Fatalf("Create() error = %v, want passenger count mismatch", err)
	}
}

func TestValidateTripOptionRejectsCorruptedCachedTotals(t *testing.T) {
	result := cachedRoundTrip(2)
	option := result.Items[0]
	option.Price++
	_, _, err := validateTripOption("search-1", option, 2)
	if err == nil {
		t.Fatal("validateTripOption() unexpectedly accepted corrupted total")
	}
}

func TestListNormalizesPaginationBeforeRepository(t *testing.T) {
	repo := &fakeOrderRepository{}
	svc := &Service{repo: repo}
	if _, err := svc.List(context.Background(), 7, orders.ListFilter{Limit: 100, Offset: -2}); err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if repo.listFilter.Limit != orders.DefaultListLimit || repo.listFilter.Offset != 0 {
		t.Fatalf("repository filter = %+v", repo.listFilter)
	}
}

func TestCleanupExpiredUsesRetentionAndNormalizedBatch(t *testing.T) {
	now := time.Date(2026, time.September, 6, 12, 0, 0, 0, time.UTC)
	repo := &fakeOrderRepository{deleted: 3}
	svc := &Service{repo: repo, now: func() time.Time { return now }}

	deleted, err := svc.CleanupExpired(context.Background(), 0)
	if err != nil {
		t.Fatalf("CleanupExpired() error = %v", err)
	}
	if deleted != 3 {
		t.Fatalf("deleted = %d, want 3", deleted)
	}
	if repo.deleteLimit != orders.DefaultCleanupLimit {
		t.Fatalf("cleanup limit = %d, want %d", repo.deleteLimit, orders.DefaultCleanupLimit)
	}
	wantBefore := now.Add(-orders.ExpiredOrderRetention)
	if !repo.deleteBefore.Equal(wantBefore) {
		t.Fatalf("cleanup before = %s, want %s", repo.deleteBefore, wantBefore)
	}
}

func TestPersistOrderRetriesOrderNumberConflict(t *testing.T) {
	repo := &fakeOrderRepository{
		createErrors: []error{orders.ErrOrderNumberConflict, orders.ErrOrderNumberConflict},
	}
	svc := &Service{repo: repo}

	order, err := svc.persistOrder(context.Background(), orders.CreateOrderParams{})
	if err != nil {
		t.Fatalf("persistOrder() error = %v", err)
	}
	if repo.createCalls != 3 {
		t.Fatalf("create calls = %d, want 3", repo.createCalls)
	}
	if order.OrderNumber == "" || order.OrderNumber != repo.created.OrderNumber {
		t.Fatalf("order number = %q, params = %q", order.OrderNumber, repo.created.OrderNumber)
	}
}

func testOrderService(result *search.CachedResult) *Service {
	return &Service{
		repo: &fakeOrderRepository{}, search: &fakeSearchReader{result: result},
		documents: &fakeDocumentValidator{}, passengers: &fakePassengerReader{},
		bonus: &fakeBonusReader{}, now: time.Now,
	}
}

func cachedRoundTrip(passengerCount int) *search.CachedResult {
	policy, _ := fare.Policy(fare.Flexible)
	outbound := search.Offer{
		ID: "outbound", Transport: "avia", PricePerPassenger: 1000, Price: 1000 * passengerCount,
		Segments: []search.Segment{{
			Order: 1, FromCityID: 1, ToCityID: 2, DepartureTime: "2026-09-10T10:00:00Z",
			ArrivalTime: "2026-09-10T12:00:00Z", RouteNumber: "SU 10001",
		}},
	}
	returnOffer := search.Offer{
		ID: "return", Transport: "avia", PricePerPassenger: 2000, Price: 2000 * passengerCount,
		Segments: []search.Segment{{
			Order: 1, FromCityID: 2, ToCityID: 1, DepartureTime: "2026-09-20T10:00:00Z",
			ArrivalTime: "2026-09-20T12:00:00Z", RouteNumber: "SU 20002",
		}},
	}
	return &search.CachedResult{
		SearchID: "search-1", Input: search.SearchInput{Passengers: passengerCount}, Total: 1,
		Items: []search.TripOption{{
			ID: "trip-1", ScheduleID: "schedule-1", SupplierCode: "atlas",
			SupplierOfferID: "supplier-offer-1", FareType: fare.Flexible, RefundPolicy: policy,
			Transport: "avia", Price: 3000 * passengerCount,
			PricePerPassenger: 3000, Outbound: outbound, Return: &returnOffer,
		}},
	}
}

func newPassenger(firstName string, lastName string, documentNumber string) orders.PassengerBooking {
	return orders.PassengerBooking{
		Source: passengers.SourceNew,
		Passenger: orders.PassengerSnapshot{
			FirstName: firstName, LastName: lastName, BirthDate: "1990-01-01", IsRussian: true,
		},
		Document: orders.DocumentBooking{Type: "internal_passport", Number: documentNumber},
	}
}

type fakeOrderRepository struct {
	created      orders.CreateOrderParams
	listFilter   orders.ListFilter
	deleteBefore time.Time
	deleteLimit  int
	deleted      int
	createCalls  int
	createErrors []error
}

func (f *fakeOrderRepository) Create(_ context.Context, params orders.CreateOrderParams) (*orders.Order, error) {
	f.created = params
	f.createCalls++
	if f.createCalls <= len(f.createErrors) {
		return nil, f.createErrors[f.createCalls-1]
	}
	return &orders.Order{
		ID: 1, OrderNumber: params.OrderNumber, Status: "created",
		TotalPrice: params.TotalPrice, CurrentTotalPrice: params.CurrentTotalPrice,
		PayableAmount: params.PayableAmount,
	}, nil
}

type fakeBonusReader struct {
	balance int
	err     error
}

func (f *fakeBonusReader) GetBalance(context.Context, int) (int, error) {
	return f.balance, f.err
}
func (f *fakeOrderRepository) List(_ context.Context, _ int, filter orders.ListFilter) (*orders.HistoryPage, error) {
	f.listFilter = filter
	return &orders.HistoryPage{Items: []orders.HistoryOrder{}}, nil
}

func (f *fakeOrderRepository) DeleteExpiredOrders(_ context.Context, before time.Time, limit int) (int, error) {
	f.deleteBefore = before
	f.deleteLimit = limit
	return f.deleted, nil
}

func (f *fakeOrderRepository) DeleteOrphanTrips(context.Context, time.Time, int) (int, error) {
	return 0, nil
}

type fakeSearchReader struct {
	result *search.CachedResult
	err    error
}

func (f *fakeSearchReader) GetCachedResult(context.Context, string) (*search.CachedResult, error) {
	return f.result, f.err
}

type fakeDocumentValidator struct {
	calls     int
	lastInput documents.BookingValidationInput
	err       error
}

func (f *fakeDocumentValidator) ValidateForBooking(_ context.Context, in documents.BookingValidationInput) (*documents.ValidatedBookingDocument, error) {
	f.calls++
	f.lastInput = in
	if f.err != nil {
		return nil, f.err
	}
	return &documents.ValidatedBookingDocument{
		SavedDocumentID: in.Document.ID,
		Type:            in.Document.Type, Number: in.Document.Number, ExpiresAt: in.Document.ExpiresAt,
		VerificationStatus: "verified", LastCheckedAt: "2026-08-22T12:00:00Z",
		Fingerprint: fmt.Sprintf("fingerprint-%d", f.calls),
	}, nil
}

type fakePassengerReader struct {
	passenger   *passengers.Passenger
	err         error
	ownerID     int
	passengerID int
}

func (f *fakePassengerReader) GetOwned(_ context.Context, ownerUserID int, passengerID int) (*passengers.Passenger, error) {
	f.ownerID = ownerUserID
	f.passengerID = passengerID
	return f.passenger, f.err
}
