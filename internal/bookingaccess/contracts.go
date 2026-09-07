package bookingaccess

import (
	"context"

	"github.com/Halturshik/TicketAgregator-API/internal/refunds"
)

type Service interface {
	Lookup(ctx context.Context, input LookupInput) (*PublicSummary, error)
	RequestAccess(ctx context.Context, input AccessRequestInput) (*AccessRequestOutput, error)
	ConfirmAccess(ctx context.Context, input AccessConfirmInput) (*AccessOutput, error)
	Details(ctx context.Context, userID *int, orderNumber string, accessToken string) (*BookingDetails, error)
	QuoteRefund(ctx context.Context, userID *int, orderNumber string, accessToken string, input RefundInput) (*RefundQuote, error)
	Refund(ctx context.Context, userID *int, orderNumber string, accessToken string, input RefundInput, idempotencyKey string) (*RefundResult, error)
}

type Repository interface {
	PublicByTicket(ctx context.Context, ticketNumber string) (*Booking, error)
	PublicByOrder(ctx context.Context, orderNumber string) (*Booking, error)
	GuestByLocator(ctx context.Context, locator Locator, email string) (*Booking, error)
	ByOrderNumber(ctx context.Context, orderNumber string) (*Booking, error)
}

type ChallengeStore interface {
	Request(ctx context.Context, challenge Challenge, code string) error
	Verify(ctx context.Context, challengeID string, code string) (*Challenge, error)
}

type AccessStore interface {
	Issue(ctx context.Context, orderID int) (string, error)
	Authorize(ctx context.Context, token string, orderID int) error
}

type Mailer interface {
	SendVerificationEmail(ctx context.Context, to string, code string) error
}

type CodeGenerator interface {
	GenerateVerificationCode() (string, error)
}

type RefundService interface {
	Quote(ctx context.Context, userID *int, orderID int, input refunds.Input) (*refunds.QuoteOutput, error)
	Refund(ctx context.Context, userID *int, orderID int, input refunds.Input) (*refunds.Result, error)
}
