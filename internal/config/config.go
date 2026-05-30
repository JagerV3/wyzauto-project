package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPAddr string
	DBURL    string
	CacheTTL time.Duration
}

func Load() Config {
	return Config{
		HTTPAddr: env("HTTP_ADDR", ":8080"),
		DBURL:    env("DATABASE_URL", "postgres://wyzauto:wyzauto@localhost:5432/wyzauto?sslmode=disable"),
		CacheTTL: time.Duration(envInt("CACHE_TTL_SECONDS", 300)) * time.Second,
	}
}

func env(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func envInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}
