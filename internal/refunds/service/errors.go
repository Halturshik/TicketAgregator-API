package service

import (
	"errors"
	"net/http"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
	"github.com/Halturshik/TicketAgregator-API/internal/refunds"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	errRefundNotAllowed    = apierror.New("refund_not_allowed", "Выбранные билеты нельзя вернуть", http.StatusConflict)
	errRefundStatus        = apierror.New("invalid_refund_status", "Билет уже возвращён или недоступен для возврата", http.StatusConflict)
	errSupplierUnavailable = apierror.New("supplier_unavailable", "Поставщик временно недоступен", http.StatusServiceUnavailable)
	errSupplierResponse    = apierror.New("supplier_invalid_response", "Поставщик вернул несогласованные данные", http.StatusBadGateway)
	errIdempotencyConflict = apierror.New("idempotency_conflict", "Этот Idempotency-Key уже использован для другого возврата", http.StatusConflict)
)

func mapError(err error) error {
	switch {
	case errors.Is(err, refunds.ErrNotFound), errors.Is(err, refunds.ErrTicketsMismatch):
		return apierror.ErrNotFound
	case errors.Is(err, refunds.ErrForbidden):
		return apierror.ErrForbidden
	case errors.Is(err, refunds.ErrInvalidStatus):
		return errRefundStatus
	case errors.Is(err, refunds.ErrNotAllowed):
		return errRefundNotAllowed
	case errors.Is(err, refunds.ErrIdempotencyConflict):
		return errIdempotencyConflict
	case errors.Is(err, refunds.ErrSupplierMismatch), status.Code(err) == codes.NotFound:
		return errSupplierResponse
	case errors.Is(err, refunds.ErrInvalidFinancialState):
		logger.Error("Некорректное финансовое состояние возврата: %v", err)
		return apierror.ErrInternal
	case status.Code(err) == codes.Unavailable || status.Code(err) == codes.DeadlineExceeded:
		return errSupplierUnavailable
	default:
		logger.Error("Ошибка возврата: %v", err)
		return err
	}
}
