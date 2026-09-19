package app

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestSwaggerRoutes(t *testing.T) {
	router := chi.NewRouter()
	registerSwagger(router)

	tests := []struct {
		name        string
		path        string
		status      int
		contentType string
		body        string
		location    string
	}{
		{
			name: "redirect", path: "/swagger", status: http.StatusMovedPermanently,
			location: "/swagger/",
		},
		{
			name: "ui", path: "/swagger/", status: http.StatusOK,
			contentType: "text/html", body: "Swagger UI",
		},
		{
			name: "specification", path: "/swagger/openapi.yaml", status: http.StatusOK,
			contentType: "application/yaml", body: "openapi: 3.0.3",
		},
		{
			name: "initializer", path: "/swagger/swagger-initializer.js", status: http.StatusOK,
			contentType: "application/javascript", body: `url: "/swagger/openapi.yaml"`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, test.path, nil)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			result := response.Result()
			defer result.Body.Close()
			body, err := io.ReadAll(result.Body)
			if err != nil {
				t.Fatalf("read response body: %v", err)
			}
			if result.StatusCode != test.status {
				t.Fatalf("status = %d, want %d", result.StatusCode, test.status)
			}
			if test.contentType != "" && !strings.Contains(result.Header.Get("Content-Type"), test.contentType) {
				t.Fatalf("content type = %q, want %q", result.Header.Get("Content-Type"), test.contentType)
			}
			if test.body != "" && !strings.Contains(string(body), test.body) {
				t.Fatalf("body does not contain %q", test.body)
			}
			if test.location != "" && result.Header.Get("Location") != test.location {
				t.Fatalf("location = %q, want %q", result.Header.Get("Location"), test.location)
			}
		})
	}
}
