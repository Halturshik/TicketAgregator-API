package app

import (
	"net/http"
	"strings"

	apidocs "github.com/Halturshik/TicketAgregator-API/docs"
	"github.com/go-chi/chi/v5"
	swaggerfiles "github.com/swaggo/files/v2"
)

const swaggerInitializer = `window.onload = function() {
  window.ui = SwaggerUIBundle({
    url: "/swagger/openapi.yaml",
    dom_id: "#swagger-ui",
    deepLinking: true,
    docExpansion: "list",
    persistAuthorization: true,
    presets: [
      SwaggerUIBundle.presets.apis,
      SwaggerUIStandalonePreset
    ],
    plugins: [
      SwaggerUIBundle.plugins.DownloadUrl
    ],
    layout: "StandaloneLayout"
  });
};
`

func registerSwagger(r chi.Router) {
	r.Get("/swagger", redirectToSwagger)
	r.Get("/swagger/openapi.yaml", serveOpenAPI)

	assets := http.StripPrefix("/swagger/", http.FileServer(http.FS(swaggerfiles.FS)))
	r.Get("/swagger/*", func(w http.ResponseWriter, request *http.Request) {
		if strings.TrimPrefix(request.URL.Path, "/swagger/") == "swagger-initializer.js" {
			w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
			_, _ = w.Write([]byte(swaggerInitializer))
			return
		}
		assets.ServeHTTP(w, request)
	})
}

func redirectToSwagger(w http.ResponseWriter, request *http.Request) {
	http.Redirect(w, request, "/swagger/", http.StatusMovedPermanently)
}

func serveOpenAPI(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	_, _ = w.Write(apidocs.OpenAPI)
}
