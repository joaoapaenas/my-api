package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port      string
	DBUrl     string
	Env       string
	JWTSecret string
	Timeout   time.Duration
}

func Load() (*Config, error) {
	// Load .env file if it exists, ignore error if it doesn't
	_ = godotenv.Load()

	cfg := &Config{
		Port:      getEnv("PORT", "8080"),
		DBUrl:     getEnv("DB_URL", ""),
		Env:       getEnv("ENV", "development"),
		JWTSecret: getEnv("JWT_SECRET", "super-secret-key-change-me"),
		Timeout:   5 * time.Second,
	}

	// Database Connection Logic
	if cfg.DBUrl == "" {
		if cfg.Env == "development" {
			// 2. Use relative path directly to match migration tool behavior
			// This avoids complex absolute path resolution issues on Windows
			cfg.DBUrl = "file:./dev.db?_pragma=journal_mode(DELETE)&_pragma=temp_store(MEMORY)&_pragma=mmap_size(0)"
		} else {
			return nil, fmt.Errorf("DB_URL environment variable is required in production")
		}
	}

	// Security Check for Production
	if cfg.Env == "production" && cfg.JWTSecret == "super-secret-key-change-me" {
		return nil, fmt.Errorf("JWT_SECRET environment variable is required in production")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
