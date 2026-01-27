package config

import (
	"fmt"
	"os"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	AppPort    string
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
		AppPort:    os.Getenv("APP_PORT"),
	}

	if cfg.DBHost == "" {
		return nil, fmt.Errorf("DB_HOST не указан")
	}
	if cfg.DBPort == "" {
		return nil, fmt.Errorf("DB_PORT не указан")
	}
	if cfg.DBUser == "" {
		return nil, fmt.Errorf("DB_USER не указан")
	}
	if cfg.DBPassword == "" {
		return nil, fmt.Errorf("DB_PASSWORD не указан")
	}
	if cfg.DBName == "" {
		return nil, fmt.Errorf("DB_NAME не указан")
	}
	if cfg.AppPort == "" {
		cfg.AppPort = "8080"
	}

	return cfg, nil
}
