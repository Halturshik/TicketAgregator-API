package app

import (
	authhandlers "github.com/Halturshik/TicketAgregator-API/internal/auth/handlers"
	authmiddleware "github.com/Halturshik/TicketAgregator-API/internal/auth/middleware"
	bonushandlers "github.com/Halturshik/TicketAgregator-API/internal/bonus/handlers"
	checkouthandlers "github.com/Halturshik/TicketAgregator-API/internal/checkout/handlers"
	documenthandlers "github.com/Halturshik/TicketAgregator-API/internal/documents/handlers"
	orderhandlers "github.com/Halturshik/TicketAgregator-API/internal/orders/handlers"
	passengerhandlers "github.com/Halturshik/TicketAgregator-API/internal/passengers/handlers"
	refundhandlers "github.com/Halturshik/TicketAgregator-API/internal/refunds/handlers"
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
	CheckoutHandler  *checkouthandlers.Handler
	BonusHandler     *bonushandlers.Handler
	RefundHandler    *refundhandlers.Handler
	AuthMiddleware   *authmiddleware.Middleware
}

func NewAPI(
	authHandler *authhandlers.Handler,
	userHandler *userhandlers.Handler,
	passengerHandler *passengerhandlers.Handler,
	documentHandler *documenthandlers.Handler,
	searchHandler *searchhandlers.Handler,
	orderHandler *orderhandlers.Handler,
	checkoutHandler *checkouthandlers.Handler,
	bonusHandler *bonushandlers.Handler,
	refundHandler *refundhandlers.Handler,
	authMiddleware *authmiddleware.Middleware,
) *API {
	return &API{
		AuthHandler:      authHandler,
		UserHandler:      userHandler,
		PassengerHandler: passengerHandler,
		DocumentHandler:  documentHandler,
		SearchHandler:    searchHandler,
		OrderHandler:     orderHandler,
		CheckoutHandler:  checkoutHandler,
		BonusHandler:     bonusHandler,
		RefundHandler:    refundHandler,
		AuthMiddleware:   authMiddleware,
	}
}

func (api *API) Init(r *chi.Mux) {
	h := api.AuthHandler

	r.Route("/api/auth", func(r chi.Router) {
		r.Post("/register", api.Handle(h.Register))
		r.Post("/register/confirm", api.Handle(h.ConfirmRegistration))

		r.Post("/login", api.Handle(h.LoginStart))
		r.Post("/login/confirm", api.Handle(h.LoginConfirm))

		r.Post("/refresh", api.Handle(h.Refresh))
		r.Post("/logout", api.Handle(h.Logout))

		r.Post("/password/forgot", api.Handle(h.ForgotPassword))
		r.Post("/password/verify-code", api.Handle(h.VerifyResetCode))
		r.Post("/password/reset", api.Handle(h.ResetPassword))
	})

	r.Route("/api/profile", func(r chi.Router) {
		r.Use(api.AuthMiddleware.Require)
		r.Get("/", api.Handle(api.UserHandler.GetProfile))
		r.Put("/", api.Handle(api.UserHandler.UpdateProfile))
	})

	r.Route("/api/passengers", func(r chi.Router) {
		r.Use(api.AuthMiddleware.Require)
		r.Get("/", api.Handle(api.PassengerHandler.List))
		r.Put("/{id}", api.Handle(api.PassengerHandler.Update))
		r.Delete("/{id}", api.Handle(api.PassengerHandler.Delete))
	})

	r.Route("/api/documents", func(r chi.Router) {
		r.Use(api.AuthMiddleware.Require)
		r.Get("/", api.Handle(api.DocumentHandler.List))
		r.Post("/", api.Handle(api.DocumentHandler.Create))
		r.Put("/{id}", api.Handle(api.DocumentHandler.Update))
		r.Delete("/{id}", api.Handle(api.DocumentHandler.Delete))
	})

	r.Route("/api/search", func(r chi.Router) {
		r.Use(api.AuthMiddleware.Optional)
		r.Get("/cities", api.Handle(api.SearchHandler.ListCities))
		r.Post("/air", api.Handle(api.SearchHandler.SearchAir))
		r.Post("/railway", api.Handle(api.SearchHandler.SearchRailway))
		r.Post("/bus", api.Handle(api.SearchHandler.SearchBus))
		r.Get("/{id}", api.Handle(api.SearchHandler.GetPage))
	})

	r.Route("/api/orders", func(r chi.Router) {
		r.Use(api.AuthMiddleware.Optional)
		r.Post("/", api.Handle(api.OrderHandler.Create))
		r.Post("/{id}/refund-quote", api.Handle(api.RefundHandler.Quote))
		r.Post("/{id}/refunds", api.Handle(api.RefundHandler.Refund))
	})

	r.Route("/api/my/orders", func(r chi.Router) {
		r.Use(api.AuthMiddleware.Require)
		r.Get("/", api.Handle(api.OrderHandler.List))
	})

	r.Route("/api/payments", func(r chi.Router) {
		r.Use(api.AuthMiddleware.Optional)
		r.Post("/mock", api.Handle(api.CheckoutHandler.Pay))
	})

	r.Route("/api/bonus", func(r chi.Router) {
		r.Use(api.AuthMiddleware.Require)
		r.Get("/history", api.Handle(api.BonusHandler.List))
	})

	//r.Get("/swagger/*", httpSwagger.Handler())

}
