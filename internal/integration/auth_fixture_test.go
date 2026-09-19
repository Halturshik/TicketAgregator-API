//go:build integration

package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/Halturshik/TicketAgregator-API/internal/app"
	"github.com/Halturshik/TicketAgregator-API/internal/auth"
	authhandlers "github.com/Halturshik/TicketAgregator-API/internal/auth/handlers"
	authmiddleware "github.com/Halturshik/TicketAgregator-API/internal/auth/middleware"
	authrepository "github.com/Halturshik/TicketAgregator-API/internal/auth/repository"
	authservice "github.com/Halturshik/TicketAgregator-API/internal/auth/service"
	authstore "github.com/Halturshik/TicketAgregator-API/internal/auth/store"
	authtoken "github.com/Halturshik/TicketAgregator-API/internal/auth/token"
	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/common/httpx"
	"github.com/alicebob/miniredis/v2"
	"github.com/go-chi/chi/v5"
	redisclient "github.com/redis/go-redis/v9"
)

type authFixture struct {
	service auth.Service
	repo    *authrepository.Repository
	tokens  *authtoken.JWTService
	mini    *miniredis.Miniredis
	mailer  *capturedMailer
	server  *httptest.Server
	client  *http.Client
}

func newAuthFixture(t *testing.T) *authFixture {
	t.Helper()
	db := openIntegrationDB(t)
	mini := miniredis.RunT(t)
	redis := redisclient.NewClient(&redisclient.Options{Addr: mini.Addr()})
	t.Cleanup(func() { _ = redis.Close() })

	repo := authrepository.NewRepository(db)
	tokens := authtoken.NewJWTService("auth-integration-secret")
	mailer := newCapturedMailer()
	service := authservice.NewService(
		repo,
		mailer,
		&sequentialCodeGenerator{},
		authstore.NewCodeService(redis),
		authstore.NewRegistrationStore(redis),
		authstore.NewLoginStore(redis),
		authstore.NewRefreshStore(redis),
		authstore.NewResetStore(redis),
		tokens,
	)
	handler := authhandlers.New(service)
	middleware := authmiddleware.New(repo, tokens)
	handlerAdapter := &app.API{}
	router := chi.NewRouter()
	router.Route("/api/auth", func(router chi.Router) {
		router.Post("/register", handlerAdapter.Handle(handler.Register))
		router.Post("/register/confirm", handlerAdapter.Handle(handler.ConfirmRegistration))
		router.Post("/login", handlerAdapter.Handle(handler.LoginStart))
		router.Post("/login/confirm", handlerAdapter.Handle(handler.LoginConfirm))
		router.Post("/refresh", handlerAdapter.Handle(handler.Refresh))
		router.Post("/logout", handlerAdapter.Handle(handler.Logout))
		router.Post("/password/forgot", handlerAdapter.Handle(handler.ForgotPassword))
		router.Post("/password/verify-code", handlerAdapter.Handle(handler.VerifyResetCode))
		router.Post("/password/reset", handlerAdapter.Handle(handler.ResetPassword))
	})
	router.With(middleware.Require).Get("/api/protected", func(w http.ResponseWriter, r *http.Request) {
		userID, ok := auth.UserIDFromContext(r.Context())
		if !ok {
			_ = httpx.WriteJSON(w, http.StatusInternalServerError, apierror.ErrInternal)
			return
		}
		_ = httpx.WriteJSON(w, http.StatusOK, map[string]int{"user_id": userID})
	})

	server := httptest.NewServer(router)
	t.Cleanup(server.Close)
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("create cookie jar: %v", err)
	}
	return &authFixture{
		service: service,
		repo:    repo,
		tokens:  tokens,
		mini:    mini,
		mailer:  mailer,
		server:  server,
		client:  &http.Client{Jar: jar},
	}
}

func (f *authFixture) post(t *testing.T, path string, payload any) authHTTPResponse {
	t.Helper()
	return f.request(t, f.client, http.MethodPost, path, payload, "")
}

func (f *authFixture) postWithRefreshToken(t *testing.T, path string, payload any, refreshToken string) authHTTPResponse {
	t.Helper()
	return f.request(t, http.DefaultClient, http.MethodPost, path, payload, refreshToken)
}

func (f *authFixture) protected(t *testing.T, accessToken string) authHTTPResponse {
	t.Helper()
	request, err := http.NewRequest(http.MethodGet, f.server.URL+"/api/protected", nil)
	if err != nil {
		t.Fatalf("create protected request: %v", err)
	}
	request.Header.Set("Authorization", "Bearer "+accessToken)
	response, err := f.client.Do(request)
	if err != nil {
		t.Fatalf("perform protected request: %v", err)
	}
	return readAuthResponse(t, response)
}

func (f *authFixture) request(
	t *testing.T,
	client *http.Client,
	method string,
	path string,
	payload any,
	refreshToken string,
) authHTTPResponse {
	t.Helper()
	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("marshal request: %v", err)
		}
		body = bytes.NewReader(data)
	}
	request, err := http.NewRequest(method, f.server.URL+path, body)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	if payload != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if refreshToken != "" {
		request.AddCookie(&http.Cookie{Name: auth.RefreshCookieName, Value: refreshToken})
	}
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	return readAuthResponse(t, response)
}

func (f *authFixture) refreshToken(t *testing.T) string {
	t.Helper()
	endpoint, err := url.Parse(f.server.URL + "/api/auth/refresh")
	if err != nil {
		t.Fatalf("parse auth URL: %v", err)
	}
	for _, cookie := range f.client.Jar.Cookies(endpoint) {
		if cookie.Name == auth.RefreshCookieName {
			return cookie.Value
		}
	}
	t.Fatal("refresh cookie is absent")
	return ""
}

type authHTTPResponse struct {
	status int
	body   []byte
	header http.Header
}

func readAuthResponse(t *testing.T, response *http.Response) authHTTPResponse {
	t.Helper()
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}
	return authHTTPResponse{status: response.StatusCode, body: body, header: response.Header.Clone()}
}

func (r authHTTPResponse) accessToken(t *testing.T) string {
	t.Helper()
	var payload struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(r.body, &payload); err != nil {
		t.Fatalf("decode access token response: %v", err)
	}
	if payload.AccessToken == "" {
		t.Fatal("access token is empty")
	}
	return payload.AccessToken
}

func (r authHTTPResponse) assertStatus(t *testing.T, want int) {
	t.Helper()
	if r.status != want {
		t.Fatalf("response status = %d, want %d; body=%s", r.status, want, r.body)
	}
}

func (r authHTTPResponse) assertAPIError(t *testing.T, want *apierror.APIError) {
	t.Helper()
	r.assertStatus(t, want.Status)
	var payload apierror.APIError
	if err := json.Unmarshal(r.body, &payload); err != nil {
		t.Fatalf("decode API error: %v", err)
	}
	if payload.Code != want.Code || payload.Message != want.Message {
		t.Fatalf("API error = code:%q message:%q, want code:%q message:%q", payload.Code, payload.Message, want.Code, want.Message)
	}
}

type sequentialCodeGenerator struct {
	sequence atomic.Int64
}

func (g *sequentialCodeGenerator) GenerateVerificationCode() (string, error) {
	return fmt.Sprintf("%06d", g.sequence.Add(1)), nil
}

type capturedMailer struct {
	mu    sync.Mutex
	codes map[string][]string
}

func newCapturedMailer() *capturedMailer {
	return &capturedMailer{codes: make(map[string][]string)}
}

func (m *capturedMailer) SendVerificationEmail(_ context.Context, email string, code string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.codes[email] = append(m.codes[email], code)
	return nil
}

func (m *capturedMailer) lastCode(t *testing.T, email string) string {
	t.Helper()
	m.mu.Lock()
	defer m.mu.Unlock()
	codes := m.codes[email]
	if len(codes) == 0 {
		t.Fatalf("verification code for %s was not sent", email)
	}
	return codes[len(codes)-1]
}

func (m *capturedMailer) count(email string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.codes[email])
}
