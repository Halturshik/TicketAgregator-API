package main

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	supplierv1 "github.com/Halturshik/TicketAgregator-API/internal/gen/supplier/v1"
	grpchealth "google.golang.org/grpc/health"
	grpchealthv1 "google.golang.org/grpc/health/grpc_health_v1"
)

type healthTestPinger struct {
	mu  sync.RWMutex
	err error
}

func (p *healthTestPinger) PingContext(context.Context) error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.err
}

func (p *healthTestPinger) setError(err error) {
	p.mu.Lock()
	p.err = err
	p.mu.Unlock()
}

func TestSupplierHealthStartsNotServing(t *testing.T) {
	healthServer := newSupplierHealthServer()
	response, err := healthServer.Check(context.Background(), &grpchealthv1.HealthCheckRequest{
		Service: supplierv1.SupplierService_ServiceDesc.ServiceName,
	})
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if response.Status != grpchealthv1.HealthCheckResponse_NOT_SERVING {
		t.Fatalf("status = %s, want NOT_SERVING", response.Status)
	}
}

func TestMonitorSupplierHealthTracksDatabaseState(t *testing.T) {
	pinger := &healthTestPinger{}
	healthServer := newSupplierHealthServer()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		monitorSupplierHealth(ctx, pinger, healthServer, 5*time.Millisecond, time.Second)
	}()

	waitForSupplierHealthStatus(t, healthServer, grpchealthv1.HealthCheckResponse_SERVING)
	pinger.setError(errors.New("database unavailable"))
	waitForSupplierHealthStatus(t, healthServer, grpchealthv1.HealthCheckResponse_NOT_SERVING)
	pinger.setError(nil)
	waitForSupplierHealthStatus(t, healthServer, grpchealthv1.HealthCheckResponse_SERVING)
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("supplier health monitor did not stop")
	}
}

func waitForSupplierHealthStatus(
	t *testing.T,
	server *grpchealth.Server,
	want grpchealthv1.HealthCheckResponse_ServingStatus,
) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		response, err := server.Check(context.Background(), &grpchealthv1.HealthCheckRequest{
			Service: supplierv1.SupplierService_ServiceDesc.ServiceName,
		})
		if err == nil && response.Status == want {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("supplier health status did not become %s", want)
}
