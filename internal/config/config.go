package config

import (
	"os"
	"time"
)

type Config struct {
	HTTPAddr       string
	DBPath         string
	SessionTTL     time.Duration
	WorkerInterval time.Duration
}

func Load() Config {
	return Config{HTTPAddr: env("HTTP_ADDR", ":8080"), DBPath: env("DB_PATH", "./dual_teacher.db"), SessionTTL: duration("SESSION_TTL", 12*time.Hour), WorkerInterval: duration("WORKER_INTERVAL", 30*time.Second)}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
func duration(key string, fallback time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	return fallback
}
