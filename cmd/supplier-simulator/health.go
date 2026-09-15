package main

import (
	"context"
	"log/slog"
	"time"

	supplierv1 "github.com/Halturshik/TicketAgregator-API/internal/gen/supplier/v1"
	grpchealth "google.golang.org/grpc/health"
	grpchealthv1 "google.golang.org/grpc/health/grpc_health_v1"
)

const (
	supplierHealthCheckInterval = 5 * time.Second
	supplierHealthCheckTimeout  = 2 * time.Second
)

type databasePinger interface {
	PingContext(context.Context) error
}

func newSupplierHealthServer() *grpchealth.Server {
	server := grpchealth.NewServer()
	setSupplierHealthStatus(server, grpchealthv1.HealthCheckResponse_NOT_SERVING)
	return server
}

func runSupplierHealth(ctx context.Context, db databasePinger, server *grpchealth.Server) {
	monitorSupplierHealth(ctx, db, server, supplierHealthCheckInterval, supplierHealthCheckTimeout)
}

func monitorSupplierHealth(
	ctx context.Context,
	db databasePinger,
	server *grpchealth.Server,
	interval time.Duration,
	timeout time.Duration,
) {
	currentStatus := grpchealthv1.HealthCheckResponse_SERVING
	check := func() {
		checkCtx, cancel := context.WithTimeout(ctx, timeout)
		err := db.PingContext(checkCtx)
		cancel()

		nextStatus := grpchealthv1.HealthCheckResponse_SERVING
		if err != nil {
			nextStatus = grpchealthv1.HealthCheckResponse_NOT_SERVING
		}
		setSupplierHealthStatus(server, nextStatus)
		if nextStatus == currentStatus {
			return
		}
		currentStatus = nextStatus
		if nextStatus == grpchealthv1.HealthCheckResponse_NOT_SERVING {
			slog.WarnContext(ctx, "PostgreSQL симулятора поставщиков недоступен", slog.Any("error", err))
			return
		}
		slog.InfoContext(ctx, "Соединение симулятора поставщиков с PostgreSQL восстановлено")
	}

	check()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			check()
		}
	}
}

func setSupplierHealthStatus(server *grpchealth.Server, status grpchealthv1.HealthCheckResponse_ServingStatus) {
	server.SetServingStatus("", status)
	server.SetServingStatus(supplierv1.SupplierService_ServiceDesc.ServiceName, status)
}
