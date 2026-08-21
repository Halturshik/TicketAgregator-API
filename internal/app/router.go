package app

import (
	"github.com/Halturshik/TicketAgregator-API/internal/auth"
	authhandlers "github.com/Halturshik/TicketAgregator-API/internal/auth/handlers"
	bonushandlers "github.com/Halturshik/TicketAgregator-API/internal/bonus/handlers"
	documenthandlers "github.com/Halturshik/TicketAgregator-API/internal/documents/handlers"
	orderhandlers "github.com/Halturshik/TicketAgregator-API/internal/orders/handlers"
	passengerhandlers "github.com/Halturshik/TicketAgregator-API/internal/passengers/handlers"
	paymenthandlers "github.com/Halturshik/TicketAgregator-API/internal/payments/handlers"
	searchhandlers "github.com/Halturshik/TicketAgregator-API/internal/search/handlers"
	userhandlers "github.com/Halturshik/TicketAgregator-API/internal/users/handlers"
	"github.com/go-chi/chi/v5"
)

type API struct {
	AuthHandler      *authhandlers.Handler
	UserHandler      *userhandlers.Handler
	PassengerHandler *passengerhandlers.Handler
	DocumentHandler  *documenthandlers.Handler
	SearchHandler    *searchhandlers.Handler
	OrderHandler     *orderhandlers.Handler
	PaymentHandler   *paymenthandlers.Handler
	BonusHandler     *bonushandlers.Handler
	AuthMiddleware   *auth.AuthMiddleware
}

func NewAPI(
	authHandler *authhandlers.Handler,
	userHandler *userhandlers.Handler,
	passengerHandler *passengerhandlers.Handler,
	documentHandler *documenthandlers.Handler,
	searchHandler *searchhandlers.Handler,
	orderHandler *orderhandlers.Handler,
	paymentHandler *paymenthandlers.Handler,
	bonusHandler *bonushandlers.Handler,
	authMiddleware *auth.AuthMiddleware,
) *API {
	return &API{
		AuthHandler:      authHandler,
		UserHandler:      userHandler,
		PassengerHandler: passengerHandler,
		DocumentHandler:  documentHandler,
		SearchHandler:    searchHandler,
		OrderHandler:     orderHandler,
		PaymentHandler:   paymentHandler,
		BonusHandler:     bonusHandler,
		AuthMiddleware:   authMiddleware,
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

	r.Route("/api/profile", func(r chi.Router) {
		r.Use(api.AuthMiddleware.Auth)
		r.Get("/", api.Handle(api.UserHandler.GetProfile))
		r.Put("/", api.Handle(api.UserHandler.UpdateProfile))
	})

	r.Route("/api/passengers", func(r chi.Router) {
		r.Use(api.AuthMiddleware.Auth)
		r.Get("/", api.Handle(api.PassengerHandler.List))
		r.Post("/", api.Handle(api.PassengerHandler.Create))
		r.Put("/{id}", api.Handle(api.PassengerHandler.Update))
	})

	r.Route("/api/documents", func(r chi.Router) {
		r.Use(api.AuthMiddleware.Auth)
		r.Get("/", api.Handle(api.DocumentHandler.List))
		r.Post("/", api.Handle(api.DocumentHandler.Create))
	})

	r.Route("/api/search", func(r chi.Router) {
		r.Get("/cities", api.Handle(api.SearchHandler.ListCities))

		r.Use(api.AuthMiddleware.OptionalAuth)
		r.Post("/air", api.Handle(api.SearchHandler.SearchAir))
		r.Post("/railway", api.Handle(api.SearchHandler.SearchRailway))
		r.Post("/bus", api.Handle(api.SearchHandler.SearchBus))
		r.Get("/{id}", api.Handle(api.SearchHandler.Page))
	})

	r.Route("/api/orders", func(r chi.Router) {
		r.Use(api.AuthMiddleware.OptionalAuth)
		r.Post("/", api.Handle(api.OrderHandler.Create))
	})

	r.Route("/api/my/orders", func(r chi.Router) {
		r.Use(api.AuthMiddleware.Auth)
		r.Get("/", api.Handle(api.OrderHandler.List))
	})

	r.Route("/api/payments", func(r chi.Router) {
		r.Use(api.AuthMiddleware.OptionalAuth)
		r.Post("/mock", api.Handle(api.PaymentHandler.Pay))
	})

	r.Route("/api/bonus", func(r chi.Router) {
		r.Use(api.AuthMiddleware.Auth)
		r.Get("/history", api.Handle(api.BonusHandler.List))
	})

	//r.Get("/swagger/*", httpSwagger.Handler())

}
