package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Halturshik/TicketAgregator-API/internal/auth"
	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
)

func TestRequire(t *testing.T) {
	tests := []struct {
		name       string
		header     string
		parseErr   error
		userErr    error
		user       *auth.UserAuth
		nilUser    bool
		wantStatus int
		wantCode   string
		wantNext   bool
		wantUserID int
	}{
		{name: "missing header", wantStatus: http.StatusUnauthorized, wantCode: apierror.ErrUnauthorized.Code},
		{name: "wrong scheme", header: "Basic token", wantStatus: http.StatusUnauthorized, wantCode: apierror.ErrInvalidTokenFormat.Code},
		{name: "missing bearer token", header: "Bearer ", wantStatus: http.StatusUnauthorized, wantCode: apierror.ErrInvalidTokenFormat.Code},
		{name: "too many bearer parts", header: "Bearer token extra", wantStatus: http.StatusUnauthorized, wantCode: apierror.ErrInvalidTokenFormat.Code},
		{name: "invalid token", header: "Bearer token", parseErr: errors.New("expired"), wantStatus: http.StatusUnauthorized, wantCode: apierror.ErrInvalidToken.Code},
		{name: "deleted user", header: "Bearer token", userErr: auth.ErrUserNotFound, wantStatus: http.StatusUnauthorized, wantCode: apierror.ErrUnauthorized.Code},
		{name: "repository failure", header: "Bearer token", userErr: errors.New("database unavailable: secret detail"), wantStatus: http.StatusInternalServerError, wantCode: apierror.ErrInternal.Code},
		{name: "nil user", header: "Bearer token", nilUser: true, wantStatus: http.StatusInternalServerError, wantCode: apierror.ErrInternal.Code},
		{name: "token version mismatch", header: "Bearer token", user: &auth.UserAuth{ID: 42, TokenVersion: 8}, wantStatus: http.StatusUnauthorized, wantCode: apierror.ErrInvalidToken.Code},
		{name: "valid token", header: "Bearer token", wantStatus: http.StatusNoContent, wantNext: true, wantUserID: 42},
		{name: "case insensitive scheme", header: "  bearer   token  ", wantStatus: http.StatusNoContent, wantNext: true, wantUserID: 42},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := &fakeTokenParser{userID: 42, version: 7, err: tt.parseErr}
			user := tt.user
			if user == nil && tt.userErr == nil && !tt.nilUser {
				user = &auth.UserAuth{ID: 42, TokenVersion: 7}
			}
			users := &fakeUserReader{user: user, err: tt.userErr}
			middleware := New(users, parser)

			called := false
			contextUserID := 0
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				contextUserID, _ = auth.UserIDFromContext(r.Context())
				w.WriteHeader(http.StatusNoContent)
			})

			request := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tt.header != "" {
				request.Header.Set("Authorization", tt.header)
			}
			response := httptest.NewRecorder()
			middleware.Require(next).ServeHTTP(response, request)

			if response.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, tt.wantStatus)
			}
			if called != tt.wantNext {
				t.Fatalf("next called = %t, want %t", called, tt.wantNext)
			}
			if tt.wantNext {
				if contextUserID != tt.wantUserID {
					t.Fatalf("context user ID = %d, want %d", contextUserID, tt.wantUserID)
				}
				return
			}

			assertAPIError(t, response, tt.wantCode)
			if tt.name == "repository failure" && strings.Contains(response.Body.String(), "secret detail") {
				t.Fatal("internal repository error leaked to response")
			}
		})
	}
}

func TestOptional(t *testing.T) {
	tests := []struct {
		name       string
		header     string
		parseErr   error
		userErr    error
		wantStatus int
		wantCode   string
		wantNext   bool
		wantUserID bool
	}{
		{name: "guest without header", wantStatus: http.StatusNoContent, wantNext: true},
		{name: "valid authenticated request", header: "Bearer token", wantStatus: http.StatusNoContent, wantNext: true, wantUserID: true},
		{name: "malformed supplied header", header: "token", wantStatus: http.StatusUnauthorized, wantCode: apierror.ErrInvalidTokenFormat.Code},
		{name: "invalid supplied token", header: "Bearer token", parseErr: errors.New("invalid"), wantStatus: http.StatusUnauthorized, wantCode: apierror.ErrInvalidToken.Code},
		{name: "repository failure", header: "Bearer token", userErr: errors.New("database unavailable"), wantStatus: http.StatusInternalServerError, wantCode: apierror.ErrInternal.Code},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := &fakeTokenParser{userID: 42, version: 7, err: tt.parseErr}
			users := &fakeUserReader{user: &auth.UserAuth{ID: 42, TokenVersion: 7}, err: tt.userErr}
			middleware := New(users, parser)

			called := false
			hasUserID := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				_, hasUserID = auth.UserIDFromContext(r.Context())
				w.WriteHeader(http.StatusNoContent)
			})

			request := httptest.NewRequest(http.MethodGet, "/optional", nil)
			if tt.header != "" {
				request.Header.Set("Authorization", tt.header)
			}
			response := httptest.NewRecorder()
			middleware.Optional(next).ServeHTTP(response, request)

			if response.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, tt.wantStatus)
			}
			if called != tt.wantNext {
				t.Fatalf("next called = %t, want %t", called, tt.wantNext)
			}
			if hasUserID != tt.wantUserID {
				t.Fatalf("context has user ID = %t, want %t", hasUserID, tt.wantUserID)
			}
			if !tt.wantNext {
				assertAPIError(t, response, tt.wantCode)
			}
		})
	}
}

func TestAuthenticationStopsBeforeUserLookupWhenTokenParsingFails(t *testing.T) {
	parser := &fakeTokenParser{err: errors.New("invalid")}
	users := &fakeUserReader{}
	middleware := New(users, parser)

	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	request.Header.Set("Authorization", "Bearer token")
	middleware.Require(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("next handler must not be called")
	})).ServeHTTP(httptest.NewRecorder(), request)

	if users.calls != 0 {
		t.Fatalf("user lookup calls = %d, want 0", users.calls)
	}
}

func assertAPIError(t *testing.T, response *httptest.ResponseRecorder, wantCode string) {
	t.Helper()
	var payload apierror.APIError
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode API error: %v", err)
	}
	if payload.Code != wantCode {
		t.Fatalf("error code = %q, want %q", payload.Code, wantCode)
	}
}

type fakeUserReader struct {
	user  *auth.UserAuth
	err   error
	calls int
}

func (f *fakeUserReader) GetUserByID(context.Context, int) (*auth.UserAuth, error) {
	f.calls++
	return f.user, f.err
}

type fakeTokenParser struct {
	userID  int
	version int
	err     error
}

func (f *fakeTokenParser) ParseAccessToken(string) (int, int, error) {
	return f.userID, f.version, f.err
}
