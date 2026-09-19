package main

import (
	"context"
	"net"
	"testing"

	"google.golang.org/grpc"
)

func TestRunSupplierServerStopsGRPCServer(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer listener.Close()

	signalCtx, stop := context.WithCancel(context.Background())
	stop()
	if err := runSupplierServer(signalCtx, listener, grpc.NewServer(), nil); err != nil {
		t.Fatalf("runSupplierServer() error = %v", err)
	}
}
