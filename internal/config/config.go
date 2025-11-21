package config

import (
	"os"
	"time"
)

type Config struct {
	Port    string
	DBUrl   string
	Env     string
	Timeout time.Duration
}

func Load() *Config {
	// In production, use a library like kelseyhightower/envconfig
	return &Config{
		Port:    getEnv("PORT", "8080"),
		DBUrl:   getEnv("DB_URL", "./dev.db"),
		Env:     getEnv("ENV", "development"),
		Timeout: 5 * time.Second,
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
