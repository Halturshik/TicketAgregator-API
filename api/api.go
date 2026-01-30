package api

import (
	"github.com/Halturshik/TicketAgregator-API/internal/auth"
	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
)


type API struct {
	AuthService *auth.Service
}

func NewAPI(store auth.UserStore, mailer auth.Mailer, redisClient *redis.Client) *API {
	return &API{
		AuthService: auth.NewService(store, mailer, redisClient),
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
