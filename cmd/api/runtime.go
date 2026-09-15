package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/platform/shutdown"
)

const (
	httpReadHeaderTimeout = 5 * time.Second
	httpReadTimeout       = 15 * time.Second
	httpWriteTimeout      = 30 * time.Second
	httpIdleTimeout       = 60 * time.Second
	shutdownTimeout       = 30 * time.Second
)

type backgroundWorker func(context.Context)

func runAPIServer(
	signalCtx context.Context,
	port string,
	handler http.Handler,
	workers ...backgroundWorker,
) error {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("запуск HTTP listener: %w", err)
	}
	defer listener.Close()

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           handler,
		ReadHeaderTimeout: httpReadHeaderTimeout,
		ReadTimeout:       httpReadTimeout,
		WriteTimeout:      httpWriteTimeout,
		IdleTimeout:       httpIdleTimeout,
	}
	workerCtx, stopWorkers := context.WithCancel(context.Background())
	defer stopWorkers()

	var runtimeWG sync.WaitGroup
	for _, worker := range workers {
		if worker == nil {
			continue
		}
		runtimeWG.Add(1)
		go func(worker backgroundWorker) {
			defer runtimeWG.Done()
			worker(workerCtx)
		}(worker)
	}

	serveErrors := make(chan error, 1)
	runtimeWG.Add(1)
	go func() {
		defer runtimeWG.Done()
		slog.Info("HTTP-сервер запущен", slog.String("address", listener.Addr().String()))
		err := server.Serve(listener)
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
		serveErrors <- err
	}()

	var serveErr error
	serveResultRead := false
	select {
	case <-signalCtx.Done():
		slog.Warn("Получен сигнал завершения, останавливаю API")
	case err := <-serveErrors:
		serveResultRead = true
		if err != nil {
			serveErr = fmt.Errorf("работа HTTP-сервера: %w", err)
		}
	}

	stopWorkers()
	shutdownErr := shutdown.New(shutdownTimeout).Run(
		shutdown.Task{
			Name:     "http_server",
			Shutdown: server.Shutdown,
			Force: func() error {
				err := server.Close()
				if errors.Is(err, http.ErrServerClosed) {
					return nil
				}
				return err
			},
		},
		shutdown.WaitGroupTask("api_runtime", &runtimeWG),
	)
	if !serveResultRead {
		select {
		case err := <-serveErrors:
			if err != nil {
				serveErr = fmt.Errorf("работа HTTP-сервера: %w", err)
			}
		default:
		}
	}
	return errors.Join(serveErr, shutdownErr)
}
