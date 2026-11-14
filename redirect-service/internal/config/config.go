package config

import "os"

type Config struct {
	URL URLConfig
}

type URLConfig struct {
	GrpcAddress string
}

// LoadConfig load the application config
func LoadConfig() *Config {
	return &Config{
		URL: URLConfig{
			GrpcAddress: os.Getenv("URL_SERVICE_GRPC_ADDRESS"),
		},
	}
}
