package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/orders"
)

const (
	orderCleanupInterval = time.Hour
	orderCleanupTimeout  = 30 * time.Second
	orderCleanupBatch    = orders.DefaultCleanupLimit
)

func runOrderCleanup(ctx context.Context, service orders.Service) {
	cleanup := func() {
		cleanupCtx, cancel := context.WithTimeout(ctx, orderCleanupTimeout)
		defer cancel()
		for {
			deleted, err := service.CleanupExpired(cleanupCtx, orderCleanupBatch)
			if err != nil {
				slog.ErrorContext(cleanupCtx, "Ошибка фоновой очистки заказов", slog.Any("error", err))
				return
			}
			if deleted < orderCleanupBatch {
				return
			}
		}
	}

	cleanup()
	ticker := time.NewTicker(orderCleanupInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cleanup()
		}
	}
}
