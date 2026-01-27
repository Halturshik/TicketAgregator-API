package database

import (
	"database/sql"
	"fmt"

	"github.com/Halturshik/TicketAgregator-API/internal/config"
	"github.com/Halturshik/TicketAgregator-API/internal/logger"
	"github.com/pressly/goose"
)

type Store struct {
	DB *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{DB: db}
}

func ConnectDB(cfg *config.Config) (*sql.DB, error) {

	connectionStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
	)

	db, err := sql.Open("postgres", connectionStr)
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к БД: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("не удалось соединиться с БД: %w", err)
	}

	logger.Info("Соединение с PostgreSQL установлено")

	if err := goose.SetDialect("postgres"); err != nil {
		return nil, fmt.Errorf("ошибка установки диалекта goose: %w", err)
	}

	migrationsDir := "./database/migrations"

	if err := goose.Up(db, migrationsDir); err != nil {
		return nil, fmt.Errorf("ошибка при применении миграций: %w", err)
	}

	logger.Info("Миграции успешно применены")

	return db, nil

}
