//go:build integration

package integration_test

import (
	"context"
	"database/sql"
	"net"
	"testing"

	supplierv1 "github.com/Halturshik/TicketAgregator-API/internal/gen/supplier/v1"
	"github.com/Halturshik/TicketAgregator-API/internal/supplier"
	suppliergrpc "github.com/Halturshik/TicketAgregator-API/internal/supplier/grpcclient"
	"github.com/Halturshik/TicketAgregator-API/internal/supplier/grpcserver"
	supplierrepository "github.com/Halturshik/TicketAgregator-API/internal/supplier/repository"
	supplierservice "github.com/Halturshik/TicketAgregator-API/internal/supplier/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

func newSupplierGateway(t *testing.T, db *sql.DB) supplier.Gateway {
	t.Helper()
	listener := bufconn.Listen(4 * 1024 * 1024)
	server := grpc.NewServer()
	supplierv1.RegisterSupplierServiceServer(server, grpcserver.New(
		supplierservice.NewService(supplierrepository.NewRepository(db)),
	))
	go func() { _ = server.Serve(listener) }()

	dialer := func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}
	connection, err := grpc.NewClient(
		"passthrough:///supplier-simulator",
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		server.Stop()
		_ = listener.Close()
		t.Fatalf("create supplier gRPC client: %v", err)
	}
	t.Cleanup(func() {
		_ = connection.Close()
		server.Stop()
		_ = listener.Close()
	})
	return suppliergrpc.New(supplierv1.NewSupplierServiceClient(connection))
}
