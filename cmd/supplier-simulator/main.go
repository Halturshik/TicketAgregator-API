package main

import (
	"net"
	"os"
	"os/signal"
	"syscall"

	supplierv1 "github.com/Halturshik/TicketAgregator-API/internal/gen/supplier/v1"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/config"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/postgres"
	"github.com/Halturshik/TicketAgregator-API/internal/supplier/grpcserver"
	supplierrepo "github.com/Halturshik/TicketAgregator-API/internal/supplier/repository"
	supplierservice "github.com/Halturshik/TicketAgregator-API/internal/supplier/service"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func main() {
	if err := godotenv.Load(); err != nil {
		logger.Warn(".env файл не найден, будут использоваться переменные окружения")
	}
	cfg, err := config.LoadSupplierConfig()
	if err != nil {
		logger.Error("Ошибка загрузки конфигурации симулятора поставщиков: %v", err)
		return
	}
	db, err := postgres.ConnectSupplierDB(cfg)
	if err != nil {
		logger.Error("Ошибка подключения симулятора поставщиков к БД: %v", err)
		return
	}
	defer db.Close()

	listener, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		logger.Error("Ошибка запуска gRPC-сервера симулятора поставщиков: %v", err)
		return
	}
	server := grpc.NewServer()
	supplierService := supplierservice.NewService(supplierrepo.NewRepository(db))
	supplierv1.RegisterSupplierServiceServer(server, grpcserver.New(supplierService))

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	go func() {
		logger.Info("Симулятор поставщиков gRPC запущен на порту %s", cfg.GRPCPort)
		if err := server.Serve(listener); err != nil {
			logger.Error("Ошибка работы симулятора поставщиков: %v", err)
		}
	}()

	<-stop
	logger.Warn("Останавливаю симулятор поставщиков")
	server.GracefulStop()
}
