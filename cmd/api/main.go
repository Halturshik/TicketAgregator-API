package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/app"
	"github.com/Halturshik/TicketAgregator-API/internal/auth"
	"github.com/Halturshik/TicketAgregator-API/internal/auth/code"
	authhandlers "github.com/Halturshik/TicketAgregator-API/internal/auth/handlers"
	authrepo "github.com/Halturshik/TicketAgregator-API/internal/auth/repository"
	authservice "github.com/Halturshik/TicketAgregator-API/internal/auth/service"
	"github.com/Halturshik/TicketAgregator-API/internal/auth/store"
	"github.com/Halturshik/TicketAgregator-API/internal/auth/token"
	bonushandlers "github.com/Halturshik/TicketAgregator-API/internal/bonus/handlers"
	bonusrepo "github.com/Halturshik/TicketAgregator-API/internal/bonus/repository"
	bonusservice "github.com/Halturshik/TicketAgregator-API/internal/bonus/service"
	documenthandlers "github.com/Halturshik/TicketAgregator-API/internal/documents/handlers"
	documentrepo "github.com/Halturshik/TicketAgregator-API/internal/documents/repository"
	documentservice "github.com/Halturshik/TicketAgregator-API/internal/documents/service"
	orderhandlers "github.com/Halturshik/TicketAgregator-API/internal/orders/handlers"
	orderrepo "github.com/Halturshik/TicketAgregator-API/internal/orders/repository"
	orderservice "github.com/Halturshik/TicketAgregator-API/internal/orders/service"
	passengerhandlers "github.com/Halturshik/TicketAgregator-API/internal/passengers/handlers"
	passengerrepo "github.com/Halturshik/TicketAgregator-API/internal/passengers/repository"
	passengerservice "github.com/Halturshik/TicketAgregator-API/internal/passengers/service"
	paymenthandlers "github.com/Halturshik/TicketAgregator-API/internal/payments/handlers"
	paymentrepo "github.com/Halturshik/TicketAgregator-API/internal/payments/repository"
	paymentservice "github.com/Halturshik/TicketAgregator-API/internal/payments/service"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/config"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/mailer"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/postgres"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/redis"
	searchhandlers "github.com/Halturshik/TicketAgregator-API/internal/search/handlers"
	searchrepo "github.com/Halturshik/TicketAgregator-API/internal/search/repository"
	searchservice "github.com/Halturshik/TicketAgregator-API/internal/search/service"
	searchstore "github.com/Halturshik/TicketAgregator-API/internal/search/store"
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
	jwtService := token.NewJWTService(cfg.JWTSecret)

	mailer := &mailer.ConsoleMailer{}
	codeGenerator := &code.RandomCodeGenerator{}
	codeStore := store.NewCodeService(redisClient)
	registrationStore := store.NewRegistrationStore(redisClient)
	loginStore := store.NewLoginStore(redisClient)
	refreshStore := store.NewRefreshStore(redisClient)
	resetPasswordStore := store.NewResetStore(redisClient)

	authService := authservice.NewService(
		authRepo,
		mailer,
		codeGenerator,
		codeStore,
		registrationStore,
		loginStore,
		refreshStore,
		resetPasswordStore,
		jwtService,
	)

	authHandler := authhandlers.New(authService)
	authMiddleware := auth.NewAuthMiddleware(authRepo, jwtService)

	userHandler := userhandlers.New(userservice.NewService(userrepo.NewRepository(infraStore.DB)))
	passengerHandler := passengerhandlers.New(passengerservice.NewService(passengerrepo.NewRepository(infraStore.DB)))
	documentSvc := documentservice.NewService(documentrepo.NewRepository(infraStore.DB))
	documentHandler := documenthandlers.New(documentSvc)

	searchSvc := searchservice.NewService(
		searchrepo.NewRepository(infraStore.DB),
		searchstore.New(redisClient),
	)
	searchHandler := searchhandlers.New(searchSvc)

	orderSvc := orderservice.NewService(orderrepo.NewRepository(infraStore.DB), searchSvc, documentSvc)
	orderHandler := orderhandlers.New(orderSvc)

	paymentHandler := paymenthandlers.New(paymentservice.NewService(paymentrepo.NewRepository(infraStore.DB)))
	bonusHandler := bonushandlers.New(bonusservice.NewService(bonusrepo.NewRepository(infraStore.DB)))

	apiServer := app.NewAPI(
		authHandler,
		userHandler,
		passengerHandler,
		documentHandler,
		searchHandler,
		orderHandler,
		paymentHandler,
		bonusHandler,
		authMiddleware,
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Ошибка при остановке сервера: %v", err)
	} else {
		logger.Info("Сервер успешно остановлен")
	}
}
