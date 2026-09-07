package app

import (
	authhandlers "github.com/Halturshik/TicketAgregator-API/internal/auth/handlers"
	authmiddleware "github.com/Halturshik/TicketAgregator-API/internal/auth/middleware"
	bonushandlers "github.com/Halturshik/TicketAgregator-API/internal/bonus/handlers"
	bookingaccesshandlers "github.com/Halturshik/TicketAgregator-API/internal/bookingaccess/handlers"
	checkouthandlers "github.com/Halturshik/TicketAgregator-API/internal/checkout/handlers"
	documenthandlers "github.com/Halturshik/TicketAgregator-API/internal/documents/handlers"
	orderhandlers "github.com/Halturshik/TicketAgregator-API/internal/orders/handlers"
	passengerhandlers "github.com/Halturshik/TicketAgregator-API/internal/passengers/handlers"
	refundhandlers "github.com/Halturshik/TicketAgregator-API/internal/refunds/handlers"
	searchhandlers "github.com/Halturshik/TicketAgregator-API/internal/search/handlers"
	triphandlers "github.com/Halturshik/TicketAgregator-API/internal/trips/handlers"
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
	BookingHandler   *bookingaccesshandlers.Handler
	TripHandler      *triphandlers.Handler
	AuthMiddleware   *authmiddleware.Middleware
	RateLimit        *RateLimitMiddleware
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
	bookingHandler *bookingaccesshandlers.Handler,
	tripHandler *triphandlers.Handler,
	authMiddleware *authmiddleware.Middleware,
	rateLimit *RateLimitMiddleware,
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
		BookingHandler:   bookingHandler,
		TripHandler:      tripHandler,
		AuthMiddleware:   authMiddleware,
		RateLimit:        rateLimit,
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
		r.With(api.RateLimit.Search).Post("/air", api.Handle(api.SearchHandler.SearchAir))
		r.With(api.RateLimit.Search).Post("/railway", api.Handle(api.SearchHandler.SearchRailway))
		r.With(api.RateLimit.Search).Post("/bus", api.Handle(api.SearchHandler.SearchBus))
		r.With(api.RateLimit.SearchPage).Get("/{id}", api.Handle(api.SearchHandler.GetPage))
	})

	r.Route("/api/orders", func(r chi.Router) {
		r.Use(api.AuthMiddleware.Optional)
		r.Post("/", api.Handle(api.OrderHandler.Create))
		r.Post("/{id}/refund-quote", api.Handle(api.RefundHandler.Quote))
		r.Post("/{id}/refunds", api.Handle(api.RefundHandler.Refund))
	})

	r.Route("/api/bookings", func(r chi.Router) {
		r.Use(api.AuthMiddleware.Optional)
		r.With(api.RateLimit.PublicLookup).Post("/lookup", api.Handle(api.BookingHandler.Lookup))
		r.With(api.RateLimit.PublicLookup).Post("/access/request", api.Handle(api.BookingHandler.RequestAccess))
		r.With(api.RateLimit.PublicLookup).Post("/access/confirm", api.Handle(api.BookingHandler.ConfirmAccess))
		r.Get("/{orderNumber}", api.Handle(api.BookingHandler.Details))
		r.Post("/{orderNumber}/refund-quote", api.Handle(api.BookingHandler.QuoteRefund))
		r.Post("/{orderNumber}/refunds", api.Handle(api.BookingHandler.Refund))
	})

	r.With(api.RateLimit.PublicLookup).Post("/api/trips/lookup", api.Handle(api.TripHandler.Lookup))

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
