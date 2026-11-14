package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	URL       URLConfig
	RateLimit RateLimitConfig
}

type URLConfig struct {
	GrpcAddress string
}

// RateLimitConfig is the per-client-IP limit applied to the public API.
type RateLimitConfig struct {
	RPS   float64
	Burst int
}

// LoadConfig load the application config
func LoadConfig() *Config {
	return &Config{
		URL: URLConfig{
			GrpcAddress: os.Getenv("URL_SERVICE_GRPC_ADDRESS"),
		},
		RateLimit: RateLimitConfig{
			RPS:   getEnvFloat("RATE_LIMIT_RPS", 10),
			Burst: getEnvInt("RATE_LIMIT_BURST", 20),
		},
	}
}

// getEnvFloat returns key parsed as a positive float, or fallback when unset or invalid.
func getEnvFloat(key string, fallback float64) float64 {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}

	value, err := strconv.ParseFloat(raw, 64)
	if err != nil || value <= 0 {
		log.Printf("config: invalid %s=%q, using %v", key, raw, fallback)
		return fallback
	}

	return value
}

// getEnvInt returns key parsed as a positive int, or fallback when unset or invalid.
func getEnvInt(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		log.Printf("config: invalid %s=%q, using %d", key, raw, fallback)
		return fallback
	}

	return value
}
