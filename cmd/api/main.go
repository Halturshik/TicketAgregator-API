package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/app"
	authhandlers "github.com/Halturshik/TicketAgregator-API/internal/auth/handlers"
	authmiddleware "github.com/Halturshik/TicketAgregator-API/internal/auth/middleware"
	authrepo "github.com/Halturshik/TicketAgregator-API/internal/auth/repository"
	authservice "github.com/Halturshik/TicketAgregator-API/internal/auth/service"
	authstore "github.com/Halturshik/TicketAgregator-API/internal/auth/store"
	authtoken "github.com/Halturshik/TicketAgregator-API/internal/auth/token"
	bonushandlers "github.com/Halturshik/TicketAgregator-API/internal/bonus/handlers"
	bonusrepo "github.com/Halturshik/TicketAgregator-API/internal/bonus/repository"
	bonusservice "github.com/Halturshik/TicketAgregator-API/internal/bonus/service"
	bookingaccesshandlers "github.com/Halturshik/TicketAgregator-API/internal/bookingaccess/handlers"
	bookingaccessrepo "github.com/Halturshik/TicketAgregator-API/internal/bookingaccess/repository"
	bookingaccessservice "github.com/Halturshik/TicketAgregator-API/internal/bookingaccess/service"
	bookingaccessstore "github.com/Halturshik/TicketAgregator-API/internal/bookingaccess/store"
	checkouthandlers "github.com/Halturshik/TicketAgregator-API/internal/checkout/handlers"
	checkoutrepo "github.com/Halturshik/TicketAgregator-API/internal/checkout/repository"
	checkoutservice "github.com/Halturshik/TicketAgregator-API/internal/checkout/service"
	documenthandlers "github.com/Halturshik/TicketAgregator-API/internal/documents/handlers"
	documentrepo "github.com/Halturshik/TicketAgregator-API/internal/documents/repository"
	documentservice "github.com/Halturshik/TicketAgregator-API/internal/documents/service"
	orderhandlers "github.com/Halturshik/TicketAgregator-API/internal/orders/handlers"
	orderrepo "github.com/Halturshik/TicketAgregator-API/internal/orders/repository"
	orderservice "github.com/Halturshik/TicketAgregator-API/internal/orders/service"
	passengerhandlers "github.com/Halturshik/TicketAgregator-API/internal/passengers/handlers"
	passengerrepo "github.com/Halturshik/TicketAgregator-API/internal/passengers/repository"
	passengerservice "github.com/Halturshik/TicketAgregator-API/internal/passengers/service"
	paymentprovider "github.com/Halturshik/TicketAgregator-API/internal/payments/provider"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/codegen"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/config"
	platformlogger "github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/mailer"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/postgres"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/ratelimit"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/redis"
	refundhandlers "github.com/Halturshik/TicketAgregator-API/internal/refunds/handlers"
	refundrepo "github.com/Halturshik/TicketAgregator-API/internal/refunds/repository"
	refundservice "github.com/Halturshik/TicketAgregator-API/internal/refunds/service"
	searchhandlers "github.com/Halturshik/TicketAgregator-API/internal/search/handlers"
	searchrepo "github.com/Halturshik/TicketAgregator-API/internal/search/repository"
	searchservice "github.com/Halturshik/TicketAgregator-API/internal/search/service"
	searchstore "github.com/Halturshik/TicketAgregator-API/internal/search/store"
	"github.com/Halturshik/TicketAgregator-API/internal/supplier"
	suppliergrpc "github.com/Halturshik/TicketAgregator-API/internal/supplier/grpcclient"
	triphandlers "github.com/Halturshik/TicketAgregator-API/internal/trips/handlers"
	triprepo "github.com/Halturshik/TicketAgregator-API/internal/trips/repository"
	tripservice "github.com/Halturshik/TicketAgregator-API/internal/trips/service"
	userhandlers "github.com/Halturshik/TicketAgregator-API/internal/users/handlers"
	userrepo "github.com/Halturshik/TicketAgregator-API/internal/users/repository"
	userservice "github.com/Halturshik/TicketAgregator-API/internal/users/service"
	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

func main() {
	if err := run(); err != nil {
		slog.Error("Приложение завершено с ошибкой", slog.Any("error", err))
		os.Exit(1)
	}
}

func run() error {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("загрузка .env: %w", err)
	}
	signalCtx, stopSignals := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopSignals()

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("загрузка конфигурации: %w", err)
	}
	log, err := platformlogger.New(platformlogger.Options{
		Service: "ticket-api",
		Format:  cfg.LogFormat,
		Level:   cfg.LogLevel,
	})
	if err != nil {
		return fmt.Errorf("инициализация логирования: %w", err)
	}
	slog.SetDefault(log)

	dbConnection, err := postgres.ConnectDB(cfg)
	if err != nil {
		return fmt.Errorf("инициализация PostgreSQL: %w", err)
	}
	defer dbConnection.Close()

	redisClient, err := redis.RedisConnection(cfg)
	if err != nil {
		return fmt.Errorf("инициализация Redis: %w", err)
	}
	defer redisClient.Close()

	infraStore := postgres.NewStore(dbConnection)
	authRepo := authrepo.NewRepository(infraStore.DB)
	jwtService := authtoken.NewJWTService(cfg.JWTSecret)

	emailMailer := &mailer.ConsoleMailer{}
	codeGenerator := codegen.NewNumericGenerator(codegen.DefaultNumericLength)
	codeStore := authstore.NewCodeService(redisClient)
	registrationStore := authstore.NewRegistrationStore(redisClient)
	loginStore := authstore.NewLoginStore(redisClient)
	refreshStore := authstore.NewRefreshStore(redisClient)
	resetPasswordStore := authstore.NewResetStore(redisClient)

	authService := authservice.NewService(
		authRepo,
		emailMailer,
		codeGenerator,
		codeStore,
		registrationStore,
		loginStore,
		refreshStore,
		resetPasswordStore,
		jwtService,
	)

	authHandler := authhandlers.New(authService)
	authMiddleware := authmiddleware.New(authRepo, jwtService)

	userHandler := userhandlers.New(userservice.NewService(userrepo.NewRepository(infraStore.DB)))
	passengerSvc := passengerservice.NewService(passengerrepo.NewRepository(infraStore.DB))
	passengerHandler := passengerhandlers.New(passengerSvc)
	documentSvc := documentservice.NewService(
		documentrepo.NewRepository(infraStore.DB),
		cfg.DocumentVerificationSecret,
	)
	documentHandler := documenthandlers.New(documentSvc)

	searchRepository := searchrepo.NewRepository(infraStore.DB)
	bootstrapCtx, bootstrapCancel := context.WithTimeout(signalCtx, 5*time.Second)
	carriers, err := searchRepository.ListCarriers(bootstrapCtx)
	bootstrapCancel()
	if err != nil {
		return fmt.Errorf("загрузка перевозчиков: %w", err)
	}
	supplierClient, supplierConnection, err := suppliergrpc.Dial(cfg.SupplierGRPCAddress)
	if err != nil {
		return fmt.Errorf("инициализация gRPC-клиента поставщиков: %w", err)
	}
	defer supplierConnection.Close()
	searchSvc, err := searchservice.NewService(
		searchRepository,
		searchstore.New(redisClient),
		supplierClient,
		supplier.DefaultProviders,
		carriers,
	)
	if err != nil {
		return fmt.Errorf("инициализация поиска: %w", err)
	}
	searchHandler := searchhandlers.New(searchSvc)

	bonusRepository := bonusrepo.NewRepository(infraStore.DB)
	orderSvc := orderservice.NewService(
		orderrepo.NewRepository(infraStore.DB), searchSvc, documentSvc, passengerSvc, bonusRepository,
	)
	orderHandler := orderhandlers.New(orderSvc)

	checkoutSvc := checkoutservice.NewService(
		checkoutrepo.NewRepository(infraStore.DB),
		paymentprovider.NewMock(),
		passengerSvc,
	)
	checkoutHandler := checkouthandlers.New(checkoutSvc)
	bonusHandler := bonushandlers.New(bonusservice.NewService(bonusRepository))
	refundSvc := refundservice.NewService(
		refundrepo.NewRepository(infraStore.DB),
		supplierClient,
	)
	refundHandler := refundhandlers.New(refundSvc)
	bookingAccessSvc := bookingaccessservice.NewService(
		bookingaccessrepo.NewRepository(infraStore.DB),
		bookingaccessstore.NewChallengeStore(redisClient),
		bookingaccessstore.NewAccessStore(redisClient),
		emailMailer,
		codeGenerator,
		refundSvc,
	)
	bookingHandler := bookingaccesshandlers.New(bookingAccessSvc)
	tripHandler := triphandlers.New(tripservice.NewService(triprepo.NewRepository(infraStore.DB)))
	rateLimitMiddleware := app.NewRateLimitMiddleware(ratelimit.New(redisClient))

	apiServer := app.NewAPI(
		authHandler,
		userHandler,
		passengerHandler,
		documentHandler,
		searchHandler,
		orderHandler,
		checkoutHandler,
		bonusHandler,
		refundHandler,
		bookingHandler,
		tripHandler,
		authMiddleware,
		rateLimitMiddleware,
	)

	r := chi.NewRouter()
	r.Use(platformlogger.HTTPMiddleware)
	apiServer.Init(r)
	healthHandler := newHealthHandler(dbConnection, redisClient, supplierConnection)
	r.Get("/health/live", healthHandler.Live)
	r.Get("/health/ready", healthHandler.Ready)

	return runAPIServer(
		signalCtx,
		cfg.AppPort,
		r,
		func(ctx context.Context) { runOrderCleanup(ctx, orderSvc) },
		func(ctx context.Context) { runRefundReconciliation(ctx, refundSvc) },
	)
}
