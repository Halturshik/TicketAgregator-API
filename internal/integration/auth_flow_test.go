//go:build integration

package integration_test

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/auth"
	"github.com/Halturshik/TicketAgregator-API/internal/auth/password"
	authstore "github.com/Halturshik/TicketAgregator-API/internal/auth/store"
	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
)

func TestCompleteAuthHTTPFlowOnPostgres(t *testing.T) {
	fixture := newAuthFixture(t)
	email := "traveler@example.com"
	oldPassword := "Password123"
	newPassword := "NewPassword456"

	registration := auth.RegisterInput{
		FirstName: "  Иван ", LastName: " Петров  ", BirthDate: "1990-01-01",
		Email: " Traveler@Example.com ", Password: oldPassword, IsRussian: true,
	}
	fixture.post(t, "/api/auth/register", registration).assertStatus(t, http.StatusOK)
	registrationCode := fixture.mailer.lastCode(t, email)

	fixture.post(t, "/api/auth/register/confirm", auth.ConfirmRegisterInput{
		Email: email, Code: "000000",
	}).assertAPIError(t, apierror.ErrInvalidVerificationCode)
	confirmation := fixture.post(t, "/api/auth/register/confirm", auth.ConfirmRegisterInput{
		Email: email, Code: registrationCode,
	})
	confirmation.assertStatus(t, http.StatusCreated)
	registrationAccess := confirmation.accessToken(t)
	assertRefreshCookieAttributes(t, confirmation.header)
	registrationRefresh := fixture.refreshToken(t)

	fixture.protected(t, registrationAccess).assertStatus(t, http.StatusOK)
	fixture.protected(t, registrationRefresh).assertAPIError(t, apierror.ErrInvalidToken)
	fixture.postWithRefreshToken(t, "/api/auth/refresh", nil, registrationAccess).
		assertAPIError(t, apierror.ErrInvalidToken)

	var storedFirstName, storedLastName, storedEmail, storedHash string
	var storedVersion int
	err := fixture.repo.DB.QueryRow(`
		SELECT first_name, last_name, email, password_hash, token_version
		FROM users WHERE email = $1
	`, email).Scan(&storedFirstName, &storedLastName, &storedEmail, &storedHash, &storedVersion)
	if err != nil {
		t.Fatalf("query registered user: %v", err)
	}
	if storedFirstName != "Иван" || storedLastName != "Петров" || storedEmail != email || storedVersion != 1 {
		t.Fatalf("stored user = %q %q %q version:%d", storedFirstName, storedLastName, storedEmail, storedVersion)
	}
	if err := password.CheckPassword(storedHash, oldPassword); err != nil {
		t.Fatalf("stored password hash does not match: %v", err)
	}

	fixture.post(t, "/api/auth/register", registration).assertAPIError(t, apierror.ErrEmailIsUsed)
	fixture.post(t, "/api/auth/logout", nil).assertStatus(t, http.StatusOK)
	fixture.postWithRefreshToken(t, "/api/auth/refresh", nil, registrationRefresh).
		assertAPIError(t, apierror.ErrInvalidToken)

	fixture.mini.FastForward(authstore.CodeRateLimitWindow + time.Second)
	beforeWrongPassword := fixture.mailer.count(email)
	fixture.post(t, "/api/auth/login", auth.LoginStartInput{
		Email: email, Password: "WrongPassword123",
	}).assertAPIError(t, apierror.ErrInvalidCredentials)
	if fixture.mailer.count(email) != beforeWrongPassword {
		t.Fatal("login code was sent for an invalid password")
	}

	fixture.post(t, "/api/auth/login", auth.LoginStartInput{
		Email: email, Password: oldPassword,
	}).assertStatus(t, http.StatusOK)
	loginCode := fixture.mailer.lastCode(t, email)
	login := fixture.post(t, "/api/auth/login/confirm", auth.LoginConfirmInput{Email: email, Code: loginCode})
	login.assertStatus(t, http.StatusOK)
	loginAccess := login.accessToken(t)
	loginRefresh := fixture.refreshToken(t)

	refreshed := fixture.post(t, "/api/auth/refresh", nil)
	refreshed.assertStatus(t, http.StatusOK)
	refreshedAccess := refreshed.accessToken(t)
	refreshedToken := fixture.refreshToken(t)
	if refreshedToken == loginRefresh {
		t.Fatal("refresh rotation returned the old refresh token")
	}
	fixture.postWithRefreshToken(t, "/api/auth/refresh", nil, loginRefresh).
		assertAPIError(t, apierror.ErrInvalidToken)
	fixture.protected(t, refreshedAccess).assertStatus(t, http.StatusOK)

	fixture.mini.FastForward(authstore.CodeRateLimitWindow + time.Second)
	unknownEmail := "unknown@example.com"
	fixture.post(t, "/api/auth/password/forgot", auth.ChangePasswordStartInput{Email: unknownEmail}).
		assertStatus(t, http.StatusOK)
	if fixture.mailer.count(unknownEmail) != 0 {
		t.Fatal("password reset code was sent for an unknown email")
	}

	fixture.post(t, "/api/auth/password/forgot", auth.ChangePasswordStartInput{Email: email}).
		assertStatus(t, http.StatusOK)
	resetCode := fixture.mailer.lastCode(t, email)
	fixture.post(t, "/api/auth/password/verify-code", auth.PasswordVerifyInput{Email: email, Code: resetCode}).
		assertStatus(t, http.StatusOK)
	fixture.post(t, "/api/auth/password/reset", auth.ChangePasswordConfirmInput{
		Email: email, Password: oldPassword,
	}).assertAPIError(t, apierror.ErrSamePassword)
	reset := fixture.post(t, "/api/auth/password/reset", auth.ChangePasswordConfirmInput{
		Email: email, Password: newPassword,
	})
	reset.assertStatus(t, http.StatusOK)
	resetAccess := reset.accessToken(t)
	resetRefresh := fixture.refreshToken(t)
	if resetRefresh == refreshedToken {
		t.Fatal("password reset did not issue a new refresh token")
	}

	fixture.protected(t, loginAccess).assertAPIError(t, apierror.ErrInvalidToken)
	fixture.protected(t, refreshedAccess).assertAPIError(t, apierror.ErrInvalidToken)
	fixture.postWithRefreshToken(t, "/api/auth/refresh", nil, refreshedToken).
		assertAPIError(t, apierror.ErrInvalidToken)
	fixture.protected(t, resetAccess).assertStatus(t, http.StatusOK)

	err = fixture.repo.DB.QueryRow(`SELECT password_hash, token_version FROM users WHERE email = $1`, email).
		Scan(&storedHash, &storedVersion)
	if err != nil {
		t.Fatalf("query reset password: %v", err)
	}
	if storedVersion != 2 {
		t.Fatalf("token version after reset = %d, want 2", storedVersion)
	}
	if err := password.CheckPassword(storedHash, newPassword); err != nil {
		t.Fatalf("new password does not match stored hash: %v", err)
	}
	if err := password.CheckPassword(storedHash, oldPassword); err == nil {
		t.Fatal("old password still matches after reset")
	}
}

func assertRefreshCookieAttributes(t *testing.T, header http.Header) {
	t.Helper()
	setCookies := strings.Join(header.Values("Set-Cookie"), "; ")
	for _, expected := range []string{
		auth.RefreshCookieName + "=",
		"Path=" + auth.RefreshCookiePath,
		"HttpOnly",
		"SameSite=Lax",
	} {
		if !strings.Contains(setCookies, expected) {
			t.Fatalf("Set-Cookie does not contain %q: %s", expected, setCookies)
		}
	}
}
