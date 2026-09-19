package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	supplierv1 "github.com/Halturshik/TicketAgregator-API/internal/gen/supplier/v1"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/config"
	platformlogger "github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/postgres"
	"github.com/Halturshik/TicketAgregator-API/internal/supplier/grpcserver"
	supplierrepo "github.com/Halturshik/TicketAgregator-API/internal/supplier/repository"
	supplierservice "github.com/Halturshik/TicketAgregator-API/internal/supplier/service"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	grpchealthv1 "google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	if err := run(); err != nil {
		slog.Error("Ошибка при работе симулятора поставщиков gRPC", slog.Any("error", err))
		os.Exit(1)
	}
}

func run() error {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("загрузка .env: %w", err)
	}
	signalCtx, stopSignals := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopSignals()

	cfg, err := config.LoadSupplierConfig()
	if err != nil {
		return fmt.Errorf("загрузка конфигурации: %w", err)
	}
	log, err := platformlogger.New(platformlogger.Options{
		Service: "supplier-simulator",
		Format:  cfg.LogFormat,
		Level:   cfg.LogLevel,
	})
	if err != nil {
		return fmt.Errorf("инициализация логирования: %w", err)
	}
	slog.SetDefault(log)
	db, err := postgres.ConnectSupplierDB(cfg)
	if err != nil {
		return fmt.Errorf("инициализация PostgreSQL: %w", err)
	}
	defer db.Close()

	listener, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		return fmt.Errorf("запуск gRPC listener: %w", err)
	}
	defer listener.Close()

	server := grpc.NewServer(grpc.UnaryInterceptor(platformlogger.UnaryServerInterceptor()))
	supplierService := supplierservice.NewService(supplierrepo.NewRepository(db))
	supplierv1.RegisterSupplierServiceServer(server, grpcserver.New(supplierService))
	healthServer := newSupplierHealthServer()
	grpchealthv1.RegisterHealthServer(server, healthServer)

	return runSupplierServer(
		signalCtx,
		listener,
		server,
		healthServer.Shutdown,
		func(ctx context.Context) { runSupplierHealth(ctx, db, healthServer) },
	)
}
