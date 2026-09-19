package shutdown

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestRunWaitsForEveryTask(t *testing.T) {
	var completed atomic.Int32
	handler := New(time.Second)

	err := handler.Run(
		Task{Name: "first", Shutdown: func(context.Context) error {
			completed.Add(1)
			return nil
		}},
		Task{Name: "second", Shutdown: func(context.Context) error {
			completed.Add(1)
			return nil
		}},
	)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if completed.Load() != 2 {
		t.Fatalf("completed tasks = %d, want 2", completed.Load())
	}
}

func TestRunJoinsTaskErrorsAndForcesFailedTask(t *testing.T) {
	wantErr := errors.New("graceful failure")
	var forced atomic.Int32
	handler := New(time.Second)

	err := handler.Run(Task{
		Name: "server",
		Shutdown: func(context.Context) error {
			return wantErr
		},
		Force: func() error {
			forced.Add(1)
			return nil
		},
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("Run() error = %v, want wrapped task error", err)
	}
	if forced.Load() != 1 {
		t.Fatalf("force calls = %d, want 1", forced.Load())
	}
}

func TestRunForcesTaskAfterTimeout(t *testing.T) {
	release := make(chan struct{})
	var forced atomic.Int32
	handler := New(20 * time.Millisecond)

	err := handler.Run(Task{
		Name: "blocked_server",
		Shutdown: func(context.Context) error {
			<-release
			return nil
		},
		Force: func() error {
			if forced.Add(1) == 1 {
				close(release)
			}
			return nil
		},
	})
	if !errors.Is(err, ErrTimeout) {
		t.Fatalf("Run() error = %v, want shutdown timeout", err)
	}
	if forced.Load() != 1 {
		t.Fatalf("force calls = %d, want 1", forced.Load())
	}
}

func TestWaitGroupTaskStopsOnCompletion(t *testing.T) {
	var group sync.WaitGroup
	group.Add(1)
	go func() {
		time.Sleep(10 * time.Millisecond)
		group.Done()
	}()

	if err := New(time.Second).Run(WaitGroupTask("workers", &group)); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}
