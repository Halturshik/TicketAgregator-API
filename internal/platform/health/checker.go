package health

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

const DefaultTimeout = 2 * time.Second

var errProbeNotConfigured = errors.New("проверка зависимости не задана")

type Probe interface {
	Check(context.Context) error
}

type ProbeFunc func(context.Context) error

type Dependency struct {
	Name  string
	Probe Probe
}

type State string

const (
	StateUp   State = "up"
	StateDown State = "down"
)

type Result struct {
	State State
	Err   error
}

type Report struct {
	Ready  bool
	Checks map[string]Result
}

type Checker struct {
	timeout      time.Duration
	dependencies []Dependency
	mu           sync.Mutex
	previous     map[string]State
}

type probeResult struct {
	index int
	name  string
	err   error
}

func NewChecker(timeout time.Duration, dependencies ...Dependency) *Checker {
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	return &Checker{
		timeout:      timeout,
		dependencies: append([]Dependency(nil), dependencies...),
		previous:     make(map[string]State, len(dependencies)),
	}
}

func (f ProbeFunc) Check(ctx context.Context) error {
	return f(ctx)
}

func (c *Checker) Check(ctx context.Context) Report {
	checkCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	report := Report{
		Ready:  true,
		Checks: make(map[string]Result, len(c.dependencies)),
	}
	if len(c.dependencies) == 0 {
		return report
	}

	results := make(chan probeResult, len(c.dependencies))
	completed := make([]bool, len(c.dependencies))
	for index, dependency := range c.dependencies {
		name := dependency.Name
		if name == "" {
			name = fmt.Sprintf("dependency_%d", index)
		}
		go func(index int, name string, probe Probe) {
			if probe == nil {
				results <- probeResult{index: index, name: name, err: errProbeNotConfigured}
				return
			}
			results <- probeResult{index: index, name: name, err: probe.Check(checkCtx)}
		}(index, name, dependency.Probe)
	}

	remaining := len(c.dependencies)
	for remaining > 0 {
		select {
		case result := <-results:
			if completed[result.index] {
				continue
			}
			completed[result.index] = true
			remaining--
			report.Checks[result.name] = dependencyResult(result.err)
		case <-checkCtx.Done():
			for index, dependency := range c.dependencies {
				if completed[index] {
					continue
				}
				name := dependency.Name
				if name == "" {
					name = fmt.Sprintf("dependency_%d", index)
				}
				report.Checks[name] = dependencyResult(checkCtx.Err())
			}
			remaining = 0
		}
	}

	for _, result := range report.Checks {
		if result.State == StateDown {
			report.Ready = false
			break
		}
	}
	c.logTransitions(ctx, report)
	return report
}

func dependencyResult(err error) Result {
	if err != nil {
		return Result{State: StateDown, Err: err}
	}
	return Result{State: StateUp}
}

func (c *Checker) logTransitions(ctx context.Context, report Report) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for name, result := range report.Checks {
		previous, known := c.previous[name]
		if known && previous == result.State {
			continue
		}
		c.previous[name] = result.State
		if result.State == StateDown {
			slog.WarnContext(ctx, "Health check компонента не пройден",
				slog.String("component", name),
				slog.Any("error", result.Err),
			)
			continue
		}
		if known {
			slog.InfoContext(ctx, "Health check компонента снова пройден", slog.String("component", name))
		}
	}
}
