package shutdown

import (
	"context"
	"sync"
)

type Task struct {
	Name     string
	Shutdown func(context.Context) error
	Force    func() error
}

func WaitGroupTask(name string, group *sync.WaitGroup) Task {
	return Task{
		Name: name,
		Shutdown: func(ctx context.Context) error {
			done := make(chan struct{})
			go func() {
				group.Wait()
				close(done)
			}()

			select {
			case <-done:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		},
	}
}
