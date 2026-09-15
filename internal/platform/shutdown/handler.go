package shutdown

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"
)

const (
	DefaultTimeout   = 30 * time.Second
	forceWaitTimeout = time.Second
)

var ErrTimeout = errors.New("превышено время корректного завершения работы")

type Handler struct {
	timeout time.Duration
}

type taskResult struct {
	id  int
	err error
}

func New(timeout time.Duration) *Handler {
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	return &Handler{timeout: timeout}
}

func (h *Handler) Run(tasks ...Task) error {
	active := make(map[int]Task, len(tasks))
	results := make(chan taskResult, len(tasks))
	ctx, cancel := context.WithTimeout(context.Background(), h.timeout)
	defer cancel()

	for id, task := range tasks {
		if task.Shutdown == nil {
			continue
		}
		if task.Name == "" {
			task.Name = fmt.Sprintf("task_%d", id)
		}
		active[id] = task
		go func(id int, task Task) {
			slog.InfoContext(ctx, "Начата остановка компонента приложения", slog.String("component", task.Name))
			results <- taskResult{id: id, err: task.Shutdown(ctx)}
		}(id, task)
	}

	if len(active) == 0 {
		return nil
	}

	slog.InfoContext(ctx, "Начато корректное завершение работы приложения",
		slog.Int("components", len(active)),
		slog.Duration("timeout", h.timeout),
	)
	errs := make([]error, 0, len(active)+1)
	handleResult := func(result taskResult, forceOnError bool) {
		task, ok := active[result.id]
		if !ok {
			return
		}
		delete(active, result.id)
		if result.err == nil {
			slog.InfoContext(ctx, "Компонент приложения остановлен", slog.String("component", task.Name))
			return
		}

		err := fmt.Errorf("остановка %s: %w", task.Name, result.err)
		errs = append(errs, err)
		slog.ErrorContext(ctx, "Ошибка корректной остановки компонента приложения",
			slog.String("component", task.Name),
			slog.Any("error", result.err),
		)
		if forceOnError {
			errs = appendForceError(errs, task)
		}
	}

	for len(active) > 0 {
		select {
		case result := <-results:
			handleResult(result, true)
		case <-ctx.Done():
			timeoutErr := fmt.Errorf("%w: timeout=%s", ErrTimeout, h.timeout)
			errs = append(errs, timeoutErr)
			slog.Error("Превышено время корректного завершения работы приложения", slog.Duration("timeout", h.timeout))

			for {
				select {
				case result := <-results:
					handleResult(result, false)
				default:
					goto force
				}
			}

		force:
			for _, task := range active {
				errs = appendForceError(errs, task)
			}
			return waitAfterForce(active, results, &errs, handleResult)
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	slog.Info("Корректное завершение работы приложения выполнено успешно")
	return nil
}

func appendForceError(errs []error, task Task) []error {
	if task.Force == nil {
		return errs
	}
	slog.Warn("Выполняется принудительная остановка компонента приложения", slog.String("component", task.Name))
	if err := task.Force(); err != nil {
		slog.Error("Ошибка принудительной остановки компонента приложения",
			slog.String("component", task.Name),
			slog.Any("error", err),
		)
		return append(errs, fmt.Errorf("принудительная остановка %s: %w", task.Name, err))
	}
	return errs
}

func waitAfterForce(
	active map[int]Task,
	results <-chan taskResult,
	errs *[]error,
	handleResult func(taskResult, bool),
) error {
	timer := time.NewTimer(forceWaitTimeout)
	defer timer.Stop()
	for len(active) > 0 {
		select {
		case result := <-results:
			handleResult(result, false)
		case <-timer.C:
			for _, task := range active {
				*errs = append(*errs, fmt.Errorf("компонент не остановлен после принудительного завершения: %s", task.Name))
			}
			return errors.Join((*errs)...)
		}
	}
	return errors.Join((*errs)...)
}
