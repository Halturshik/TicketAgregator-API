package app

import (
	"github.com/Halturshik/TicketAgregator-API/internal/auth/handlers"
	"github.com/go-chi/chi/v5"
)

type API struct {
	AuthHandler *handlers.Handler
}

func NewAPI(authHandler *handlers.Handler) *API {
	return &API{
		AuthHandler: authHandler,
	}
}

func (api *API) Init(r *chi.Mux) {
	h := api.AuthHandler

	r.Route("/api/auth", func(r chi.Router) {
		r.Post("/register", api.Handle(h.RegisterHandler))
		r.Post("/register/confirm", api.Handle(h.ConfirmRegistrationHandler))

		r.Post("/login", api.Handle(h.LoginStartHandler))
		r.Post("/login/confirm", api.Handle(h.LoginConfirmHandler))

		r.Post("/refresh", api.Handle(h.Refresh))
		r.Post("/logout", api.Handle(h.LogoutHandler))

		r.Post("/password/forgot", api.Handle(h.ForgotPasswordHandler))
		r.Post("/password/verify-code", api.Handle(h.VerifyResetCodeHandler))
		r.Post("/password/reset", api.Handle(h.ResetPasswordHandler))
	})

	//r.Get("/swagger/*", httpSwagger.Handler())

}
