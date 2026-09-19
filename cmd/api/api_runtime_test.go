package main

import (
	"context"
	"strings"
	"testing"
)

func TestRunAPIServerStopsBackgroundWorkers(t *testing.T) {
	signalCtx, stop := context.WithCancel(context.Background())
	stop()
	workerStopped := make(chan struct{})

	err := runAPIServer(signalCtx, "0", nil, func(ctx context.Context) {
		<-ctx.Done()
		close(workerStopped)
	})
	if err != nil {
		t.Fatalf("runAPIServer() error = %v", err)
	}
	select {
	case <-workerStopped:
	default:
		t.Fatal("background worker was not stopped")
	}
}

func TestRunAPIServerReturnsListenerError(t *testing.T) {
	err := runAPIServer(context.Background(), "70000", nil)
	if err == nil || !strings.Contains(err.Error(), "HTTP listener") {
		t.Fatalf("runAPIServer() error = %v, want listener error", err)
	}
}
