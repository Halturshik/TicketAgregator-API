package main

import (
	"net"
	"testing"

	supplierv1 "github.com/Halturshik/TicketAgregator-API/internal/gen/supplier/v1"
	"google.golang.org/grpc"
	grpchealth "google.golang.org/grpc/health"
	grpchealthv1 "google.golang.org/grpc/health/grpc_health_v1"
)

func TestCheckSupplierAcceptsServingService(t *testing.T) {
	startHealthServer(t, grpchealthv1.HealthCheckResponse_SERVING)
	if err := checkSupplier(); err != nil {
		t.Fatalf("checkSupplier() error = %v", err)
	}
}

func TestCheckSupplierRejectsNotServingService(t *testing.T) {
	startHealthServer(t, grpchealthv1.HealthCheckResponse_NOT_SERVING)
	if err := checkSupplier(); err == nil {
		t.Fatal("checkSupplier() expected NOT_SERVING error")
	}
}

func startHealthServer(t *testing.T, status grpchealthv1.HealthCheckResponse_ServingStatus) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen for health server: %v", err)
	}
	_, port, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		listener.Close()
		t.Fatalf("parse health server address: %v", err)
	}
	t.Setenv("SUPPLIER_GRPC_PORT", port)

	server := grpc.NewServer()
	healthServer := grpchealth.NewServer()
	healthServer.SetServingStatus(supplierv1.SupplierService_ServiceDesc.ServiceName, status)
	grpchealthv1.RegisterHealthServer(server, healthServer)
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(func() {
		server.Stop()
		_ = listener.Close()
	})
}
