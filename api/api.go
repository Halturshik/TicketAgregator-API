package api

import (
	"github.com/Halturshik/TicketAgregator-API/internal/auth"
	"github.com/go-chi/chi/v5"
)

type Store interface {
}

type API struct {
	Store Store
	AuthService *auth.Service
}

func NewAPI(store Store) *API {
	return &API{
		Store:       store,
		AuthService: auth.NewService(store),
	}
}

func (api *API) Init(r *chi.Mux) {
	r.Route("/subscriptions", func(r chi.Router) {
		r.Post("/", api.)
	})

	r.Route("/users/{user_id}/subscriptions", func(r chi.Router) {
		r.Get("/", api.)
		r.Get("/{service_name}", api.)
		r.Put("/{service_name}", api.)
		r.Delete("/{service_name}", api.)
		r.Post("/{service_name}/total", api.)

	})

	//r.Get("/swagger/*", httpSwagger.Handler())

}
