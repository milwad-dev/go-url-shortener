package config

import (
	"fmt"
	"os"
)

// Config represent the application config
type Config struct {
	DB DBConfig
}

// DBConfig represent the database config
type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

// DSN return the database DSN
func (cfg *Config) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.DB.User,
		cfg.DB.Password,
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.Name,
		cfg.DB.SSLMode,
	)
}

// LoadConfig load all application config
func LoadConfig() *Config {
	return &Config{
		DB: DBConfig{
			Host:     os.Getenv("DATABASE_HOST"),
			Port:     os.Getenv("DATABASE_PORT"),
			User:     os.Getenv("DATABASE_USER"),
			Password: os.Getenv("DATABASE_PASSWORD"),
			Name:     os.Getenv("DATABASE_NAME"),
			SSLMode:  getEnv("DATABASE_SSLMODE", "disable"),
		},
	}
}

// getEnv returns the environment variable key, or fallback when it is unset or empty.
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
