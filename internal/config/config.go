package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port    string
	DBUrl   string
	Env     string
	Timeout time.Duration
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Port:    getEnv("PORT", "8080"),
		DBUrl:   getEnv("DB_URL", ""),
		Env:     getEnv("ENV", "development"),
		Timeout: 5 * time.Second,
	}

	if cfg.DBUrl == "" {
		if cfg.Env == "development" {
			// 1. Get Absolute Path
			wd, _ := os.Getwd()
			rawPath := filepath.Join(wd, "dev.db")

			// 2. Force Forward Slashes (Windows "B:\" breaks URI parsing, "B:/" works)
			cleanPath := filepath.ToSlash(rawPath)

			// 3. Construct "Compatibility Mode" DSN
			// file: prefix is required for parameters to work
			// _pragma=journal_mode(DELETE): No WAL/SHM files
			// _pragma=temp_store(MEMORY): No temp files on disk
			// _pragma=mmap_size(0): No memory mapping (fixes "Out of Memory" on some drives)
			cfg.DBUrl = fmt.Sprintf("file:%s?_pragma=journal_mode(DELETE)&_pragma=temp_store(MEMORY)&_pragma=mmap_size(0)", cleanPath)
		} else {
			return nil, fmt.Errorf("DB_URL environment variable is required")
		}
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
