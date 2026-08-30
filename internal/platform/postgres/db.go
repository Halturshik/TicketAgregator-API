package postgres

import (
	"database/sql"
	"fmt"
	"sync"

	"github.com/Halturshik/TicketAgregator-API/internal/platform/config"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
	"github.com/pressly/goose"
)

const (
	appMigrationsDir       = "./migrations"
	supplierMigrationsDir  = "./supplier-migrations"
	appMigrationTable      = "goose_db_version"
	supplierMigrationTable = "supplier_goose_db_version"
)

var migrationMu sync.Mutex

type Store struct {
	DB *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{DB: db}
}

func ConnectDB(cfg *config.Config) (*sql.DB, error) {
	db, err := openDB(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)
	if err != nil {
		return nil, err
	}
	if err := migrate(db, appMigrationsDir, appMigrationTable); err != nil {
		db.Close()
		return nil, err
	}
	logger.Info("Миграции API успешно применены")
	return db, nil
}

func ConnectSupplierDB(cfg *config.SupplierConfig) (*sql.DB, error) {
	db, err := openDB(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)
	if err != nil {
		return nil, err
	}
	if err := migrate(db, supplierMigrationsDir, supplierMigrationTable); err != nil {
		db.Close()
		return nil, err
	}
	logger.Info("Миграции симулятора поставщиков успешно применены")
	return db, nil
}

func openDB(host, port, user, password, name string) (*sql.DB, error) {
	connectionStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, name,
	)
	db, err := sql.Open("postgres", connectionStr)
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к БД: %w", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("не удалось соединиться с БД: %w", err)
	}
	logger.Info("Соединение с PostgreSQL установлено")
	return db, nil
}

func migrate(db *sql.DB, migrationsDir, tableName string) error {
	migrationMu.Lock()
	defer migrationMu.Unlock()
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("ошибка установки диалекта goose: %w", err)
	}
	goose.SetTableName(tableName)
	defer goose.SetTableName(appMigrationTable)
	if err := goose.Up(db, migrationsDir); err != nil {
		return fmt.Errorf("ошибка при применении миграций: %w", err)
	}
	return nil
}
