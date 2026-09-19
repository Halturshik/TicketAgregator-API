package health

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestCheckerRunsDependenciesConcurrently(t *testing.T) {
	started := make(chan struct{}, 2)
	release := make(chan struct{})
	probe := ProbeFunc(func(context.Context) error {
		started <- struct{}{}
		<-release
		return nil
	})
	checker := NewChecker(time.Second,
		Dependency{Name: "first", Probe: probe},
		Dependency{Name: "second", Probe: probe},
	)

	reports := make(chan Report, 1)
	go func() {
		reports <- checker.Check(context.Background())
	}()

	for range 2 {
		select {
		case <-started:
		case <-time.After(time.Second):
			t.Fatal("dependencies were not started concurrently")
		}
	}
	close(release)

	report := <-reports
	if !report.Ready {
		t.Fatalf("report.Ready = false, checks = %+v", report.Checks)
	}
}

func TestCheckerMarksFailedAndTimedOutDependenciesDown(t *testing.T) {
	wantErr := errors.New("connection details must remain internal")
	checker := NewChecker(20*time.Millisecond,
		Dependency{Name: "failed", Probe: ProbeFunc(func(context.Context) error {
			return wantErr
		})},
		Dependency{Name: "timed_out", Probe: ProbeFunc(func(ctx context.Context) error {
			<-ctx.Done()
			return ctx.Err()
		})},
	)

	report := checker.Check(context.Background())
	if report.Ready {
		t.Fatal("report.Ready = true, want false")
	}
	if result := report.Checks["failed"]; result.State != StateDown || !errors.Is(result.Err, wantErr) {
		t.Fatalf("failed result = %+v", result)
	}
	if result := report.Checks["timed_out"]; result.State != StateDown || !errors.Is(result.Err, context.DeadlineExceeded) {
		t.Fatalf("timed_out result = %+v", result)
	}
}
