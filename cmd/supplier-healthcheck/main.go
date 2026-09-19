package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	supplierv1 "github.com/Halturshik/TicketAgregator-API/internal/gen/supplier/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	grpchealthv1 "google.golang.org/grpc/health/grpc_health_v1"
)

const (
	defaultSupplierPort = "9090"
	healthCheckTimeout  = 2 * time.Second
)

func main() {
	if err := checkSupplier(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func checkSupplier() error {
	ctx, cancel := context.WithTimeout(context.Background(), healthCheckTimeout)
	defer cancel()
	port := os.Getenv("SUPPLIER_GRPC_PORT")
	if port == "" {
		port = defaultSupplierPort
	}

	connection, err := grpc.NewClient(
		net.JoinHostPort("127.0.0.1", port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("подключение к gRPC-сервису поставщиков: %w", err)
	}
	defer connection.Close()

	response, err := grpchealthv1.NewHealthClient(connection).Check(ctx, &grpchealthv1.HealthCheckRequest{
		Service: supplierv1.SupplierService_ServiceDesc.ServiceName,
	})
	if err != nil {
		return fmt.Errorf("проверка gRPC-сервиса поставщиков: %w", err)
	}
	if response.Status != grpchealthv1.HealthCheckResponse_SERVING {
		return fmt.Errorf("gRPC-сервис поставщиков не готов: status=%s", response.Status)
	}
	return nil
}
