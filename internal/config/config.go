package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	AppPort    string

	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
		AppPort:    os.Getenv("APP_PORT"),

		RedisHost:     os.Getenv("REDIS_HOST"),
		RedisPort:     os.Getenv("REDIS_PORT"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
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

	if cfg.RedisHost == "" {
		cfg.RedisHost = "localhost"
	}
	if cfg.RedisPort == "" {
		cfg.RedisPort = "6379"
	}
	if cfg.RedisPassword == "" {
		return nil, fmt.Errorf("REDIS_PASSWORD не указан")
	}
	if redisVer := os.Getenv("REDIS_DB"); redisVer != "" {
		var err error
		cfg.RedisDB, err = strconv.Atoi(redisVer)
		if err != nil {
			return nil, fmt.Errorf("REDIS_DB некорректен: %w", err)
		}
	} else {
		cfg.RedisDB = 0
	}

	return cfg, nil
}
