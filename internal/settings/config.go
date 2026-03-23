package settings

import (
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultHTTPPort      = 8080
	defaultPostgresDSN   = "postgres://postgres:postgres@localhost:5432/room_booking?sslmode=disable"
	defaultJWTSecret     = "dev-secret-change-me"
	defaultJWTTTLMinutes = 60
)

type Config struct {
	HTTP     HTTPConfig
	Postgres PostgresConfig
	JWT      JWTConfig
}

type HTTPConfig struct {
	Port int
}

type PostgresConfig struct {
	DSN string
}

type JWTConfig struct {
	Secret string
	TTL    time.Duration
}

func NewConfig() Config {
	return Config{
		HTTP: HTTPConfig{
			Port: getInt("APP_HTTP_PORT", defaultHTTPPort),
		},
		Postgres: PostgresConfig{
			DSN: getString("APP_POSTGRES_DSN", defaultPostgresDSN),
		},
		JWT: JWTConfig{
			Secret: getString("APP_JWT_SECRET", defaultJWTSecret),
			TTL:    time.Duration(getInt("APP_JWT_TTL_MINUTES", defaultJWTTTLMinutes)) * time.Minute,
		},
	}
}

func getInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}

	return parsed
}

func getString(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	return value
}
