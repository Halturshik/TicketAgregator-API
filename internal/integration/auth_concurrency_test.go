//go:build integration

package integration_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/auth"
	authstore "github.com/Halturshik/TicketAgregator-API/internal/auth/store"
	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
)

func TestConcurrentRegistrationConfirmationCreatesOneUser(t *testing.T) {
	fixture := newAuthFixture(t)
	email := "concurrent-auth@example.com"
	input := auth.RegisterInput{
		FirstName: "Иван", LastName: "Петров", BirthDate: "1990-01-01",
		Email: email, Password: "Password123", IsRussian: true,
	}
	if err := fixture.service.StartRegistration(context.Background(), input); err != nil {
		t.Fatalf("StartRegistration() error = %v", err)
	}
	code := fixture.mailer.lastCode(t, email)

	const workers = 8
	start := make(chan struct{})
	results := make(chan error, workers)
	var group sync.WaitGroup
	group.Add(workers)
	for range workers {
		go func() {
			defer group.Done()
			<-start
			_, err := fixture.service.ConfirmRegistration(context.Background(), auth.ConfirmRegisterInput{
				Email: email,
				Code:  code,
			})
			results <- err
		}()
	}
	close(start)
	group.Wait()
	close(results)

	succeeded := 0
	rejected := 0
	for err := range results {
		switch {
		case err == nil:
			succeeded++
		case errors.Is(err, apierror.ErrInvalidVerificationCode), errors.Is(err, apierror.ErrCodeExpired):
			rejected++
		default:
			t.Fatalf("unexpected confirmation error = %v", err)
		}
	}
	if succeeded != 1 || rejected != workers-1 {
		t.Fatalf("confirmation results = success:%d rejected:%d, want 1 and %d", succeeded, rejected, workers-1)
	}

	var users int
	if err := fixture.repo.DB.QueryRow(`SELECT COUNT(*) FROM users WHERE email = $1`, email).Scan(&users); err != nil {
		t.Fatalf("count users: %v", err)
	}
	if users != 1 {
		t.Fatalf("created users = %d, want 1", users)
	}
}

func TestConcurrentRefreshRotatesTokenOnceOnPostgres(t *testing.T) {
	fixture := newAuthFixture(t)
	email := "concurrent-refresh@example.com"
	input := auth.RegisterInput{
		FirstName: "Иван", LastName: "Петров", BirthDate: "1990-01-01",
		Email: email, Password: "Password123", IsRussian: true,
	}
	if err := fixture.service.StartRegistration(context.Background(), input); err != nil {
		t.Fatalf("StartRegistration() error = %v", err)
	}
	initial, err := fixture.service.ConfirmRegistration(context.Background(), auth.ConfirmRegisterInput{
		Email: email,
		Code:  fixture.mailer.lastCode(t, email),
	})
	if err != nil {
		t.Fatalf("ConfirmRegistration() error = %v", err)
	}

	const workers = 12
	start := make(chan struct{})
	type result struct {
		output *auth.LoginOutput
		err    error
	}
	results := make(chan result, workers)
	var group sync.WaitGroup
	group.Add(workers)
	for range workers {
		go func() {
			defer group.Done()
			<-start
			output, err := fixture.service.Refresh(context.Background(), initial.RefreshToken)
			results <- result{output: output, err: err}
		}()
	}
	close(start)
	group.Wait()
	close(results)

	succeeded := 0
	rejected := 0
	var winningToken string
	for item := range results {
		switch {
		case item.err == nil:
			succeeded++
			winningToken = item.output.RefreshToken
		case errors.Is(item.err, apierror.ErrInvalidToken):
			rejected++
		default:
			t.Fatalf("unexpected refresh error = %v", item.err)
		}
	}
	if succeeded != 1 || rejected != workers-1 {
		t.Fatalf("refresh results = success:%d rejected:%d, want 1 and %d", succeeded, rejected, workers-1)
	}
	if _, err := fixture.service.Refresh(context.Background(), winningToken); err != nil {
		t.Fatalf("winning refresh token is unusable: %v", err)
	}
}

func TestConcurrentPasswordResetConsumesVerificationOnceOnPostgres(t *testing.T) {
	fixture := newAuthFixture(t)
	email := "concurrent-reset@example.com"
	input := auth.RegisterInput{
		FirstName: "Иван", LastName: "Петров", BirthDate: "1990-01-01",
		Email: email, Password: "Password123", IsRussian: true,
	}
	if err := fixture.service.StartRegistration(context.Background(), input); err != nil {
		t.Fatalf("StartRegistration() error = %v", err)
	}
	if _, err := fixture.service.ConfirmRegistration(context.Background(), auth.ConfirmRegisterInput{
		Email: email,
		Code:  fixture.mailer.lastCode(t, email),
	}); err != nil {
		t.Fatalf("ConfirmRegistration() error = %v", err)
	}

	fixture.mini.FastForward(authstore.CodeRateLimitWindow + time.Second)
	if err := fixture.service.ForgotPassword(context.Background(), email); err != nil {
		t.Fatalf("ForgotPassword() error = %v", err)
	}
	if err := fixture.service.VerifyResetCode(context.Background(), auth.PasswordVerifyInput{
		Email: email,
		Code:  fixture.mailer.lastCode(t, email),
	}); err != nil {
		t.Fatalf("VerifyResetCode() error = %v", err)
	}

	passwords := []string{"NewPassword456", "AnotherPassword789"}
	start := make(chan struct{})
	results := make(chan error, len(passwords))
	var group sync.WaitGroup
	group.Add(len(passwords))
	for _, newPassword := range passwords {
		go func() {
			defer group.Done()
			<-start
			_, err := fixture.service.ResetPassword(context.Background(), auth.ChangePasswordConfirmInput{
				Email: email, Password: newPassword,
			})
			results <- err
		}()
	}
	close(start)
	group.Wait()
	close(results)

	succeeded := 0
	rejected := 0
	for err := range results {
		switch {
		case err == nil:
			succeeded++
		case errors.Is(err, apierror.ErrUnauthorized):
			rejected++
		default:
			t.Fatalf("unexpected password reset error = %v", err)
		}
	}
	if succeeded != 1 || rejected != 1 {
		t.Fatalf("password reset results = success:%d rejected:%d, want 1 and 1", succeeded, rejected)
	}

	var version int
	if err := fixture.repo.DB.QueryRow(`SELECT token_version FROM users WHERE email = $1`, email).Scan(&version); err != nil {
		t.Fatalf("query token version: %v", err)
	}
	if version != 2 {
		t.Fatalf("token version = %d, want 2", version)
	}
}
