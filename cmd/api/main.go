package main

import (
	"context"
	"net/http"
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
	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
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
	if err := godotenv.Load(); err != nil {
		logger.Warn(".env файл не найден, будут использоваться переменные окружения")
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Error("Ошибка загрузки конфигурации: %v", err)
	}

	dbConnection, err := postgres.ConnectDB(cfg)
	if err != nil {
		logger.Error("Ошибка при подключении к БД: %v", err)
	}
	defer dbConnection.Close()

	redisClient, err := redis.RedisConnection(cfg)
	if err != nil {
		logger.Error("Ошибка при подключении к Redis: %v", err)
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
	bootstrapCtx, bootstrapCancel := context.WithTimeout(context.Background(), 5*time.Second)
	carriers, err := searchRepository.ListCarriers(bootstrapCtx)
	bootstrapCancel()
	if err != nil {
		logger.Error("Ошибка загрузки перевозчиков: %v", err)
		return
	}
	supplierClient, supplierConnection, err := suppliergrpc.Dial(cfg.SupplierGRPCAddress)
	if err != nil {
		logger.Error("Ошибка инициализации gRPC-клиента поставщиков: %v", err)
		return
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
		logger.Error("Ошибка инициализации поиска: %v", err)
		return
	}
	searchHandler := searchhandlers.New(searchSvc)

	bonusRepository := bonusrepo.NewRepository(infraStore.DB)
	orderSvc := orderservice.NewService(
		orderrepo.NewRepository(infraStore.DB), searchSvc, documentSvc, passengerSvc, bonusRepository,
	)
	orderHandler := orderhandlers.New(orderSvc)
	orderCleanupCtx, stopOrderCleanup := context.WithCancel(context.Background())
	defer stopOrderCleanup()
	go runOrderCleanup(orderCleanupCtx, orderSvc)

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
	reconciliationCtx, stopReconciliation := context.WithCancel(context.Background())
	defer stopReconciliation()
	go runRefundReconciliation(reconciliationCtx, refundSvc)
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
	apiServer.Init(r)
	// r.Use(api.LoggingMiddleware)

	srv := &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: r,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		logger.Info("Сервер запущен на порту %s", cfg.AppPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Ошибка при запуске сервера: %v", err)
		}
	}()

	sig := <-stop
	logger.Warn("Получен сигнал завершения: %v, останавливаю сервер...", sig)
	stopReconciliation()
	stopOrderCleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Ошибка при остановке сервера: %v", err)
	} else {
		logger.Info("Сервер успешно остановлен")
	}
}
