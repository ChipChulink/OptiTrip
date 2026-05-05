package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort    string
	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string
	CacheTTL  int
	LogLevel   string
}

func Load() *Config {
	_ = godotenv.Load("configs/.env")

	cacheTTL := 10
	if ttl := os.Getenv("CACHE_TTL"); ttl != "" {
		if v, err := strconv.Atoi(ttl); err == nil {
			cacheTTL = v
		}
	}

	return &Config{
		AppPort:    getEnv("APP_PORT", "8080"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBName:     getEnv("DB_NAME", "optitrip"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		CacheTTL:  cacheTTL,
		LogLevel:   getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
