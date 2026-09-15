package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/platform/shutdown"
	"google.golang.org/grpc"
)

const supplierShutdownTimeout = 30 * time.Second

type supplierBackgroundWorker func(context.Context)

func runSupplierServer(
	signalCtx context.Context,
	listener net.Listener,
	server *grpc.Server,
	beforeShutdown func(),
	workers ...supplierBackgroundWorker,
) error {
	workerCtx, stopWorkers := context.WithCancel(context.Background())
	defer stopWorkers()

	serveErrors := make(chan error, 1)
	var runtimeWG sync.WaitGroup
	for _, worker := range workers {
		if worker == nil {
			continue
		}
		runtimeWG.Add(1)
		go func(worker supplierBackgroundWorker) {
			defer runtimeWG.Done()
			worker(workerCtx)
		}(worker)
	}

	runtimeWG.Add(1)
	go func() {
		defer runtimeWG.Done()
		slog.Info("Симулятор поставщиков gRPC запущен", slog.String("address", listener.Addr().String()))
		err := server.Serve(listener)
		if errors.Is(err, grpc.ErrServerStopped) {
			err = nil
		}
		serveErrors <- err
	}()

	var serveErr error
	serveResultRead := false
	select {
	case <-signalCtx.Done():
		slog.Warn("Получен сигнал завершения, останавливаю симулятор поставщиков")
	case err := <-serveErrors:
		serveResultRead = true
		if err != nil {
			serveErr = fmt.Errorf("работа gRPC-сервера: %w", err)
		}
	}

	if beforeShutdown != nil {
		beforeShutdown()
	}
	stopWorkers()
	shutdownErr := shutdown.New(supplierShutdownTimeout).Run(
		shutdown.Task{
			Name: "supplier_grpc_server",
			Shutdown: func(context.Context) error {
				server.GracefulStop()
				return nil
			},
			Force: func() error {
				server.Stop()
				return nil
			},
		},
		shutdown.WaitGroupTask("supplier_runtime", &runtimeWG),
	)
	if !serveResultRead {
		select {
		case err := <-serveErrors:
			if err != nil {
				serveErr = fmt.Errorf("работа gRPC-сервера: %w", err)
			}
		default:
		}
	}
	return errors.Join(serveErr, shutdownErr)
}
