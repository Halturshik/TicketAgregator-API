package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/refunds"
)

const (
	refundReconciliationInterval = 15 * time.Second
	refundReconciliationTimeout  = 20 * time.Second
	refundReconciliationBatch    = 10
)

func runRefundReconciliation(ctx context.Context, service refunds.Service) {
	ticker := time.NewTicker(refundReconciliationInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			reconcileCtx, cancel := context.WithTimeout(ctx, refundReconciliationTimeout)
			err := service.Reconcile(reconcileCtx, refundReconciliationBatch)
			cancel()
			if err != nil && ctx.Err() == nil {
				slog.ErrorContext(ctx, "Ошибка фоновой сверки возвратов", slog.Any("error", err))
			}
		}
	}
}
