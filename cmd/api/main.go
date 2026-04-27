package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/app"
	"github.com/Halturshik/TicketAgregator-API/internal/auth/repository"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/config"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/mailer"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/postgres"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/redis"
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
	authRepo := repository.NewRepository(infraStore.DB)

	mail := &mailer.ConsoleMailer{}
	apiServer := app.NewAPI(authRepo, mail, redisClient)

	r := chi.NewRouter()

	// r.Use(api.LoggingMiddleware)

	apiServer.Init(r)

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
