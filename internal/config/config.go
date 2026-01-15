package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port              string
	DatabasePath      string
	ScanIntervalHours int
	LogLevel          string
}

func Load() *Config {
	return &Config{
		Port:              getEnv("STALE_PORT", "3000"),
		DatabasePath:      getEnv("STALE_DB_PATH", "./stale.db"),
		ScanIntervalHours: getEnvInt("STALE_SCAN_INTERVAL", 24),
		LogLevel:          getEnv("STALE_LOG_LEVEL", "info"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}
