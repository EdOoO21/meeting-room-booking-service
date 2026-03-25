package settings

import (
	"testing"
	"time"
)

func TestNewConfig_UsesDefaultsWhenEnvIsMissing(t *testing.T) {
	t.Setenv("APP_HTTP_PORT", "")
	t.Setenv("APP_POSTGRES_DSN", "")
	t.Setenv("APP_JWT_SECRET", "")
	t.Setenv("APP_JWT_TTL_MINUTES", "")

	cfg := NewConfig()

	if cfg.HTTP.Port != defaultHTTPPort {
		t.Fatalf("HTTP.Port = %d, want %d", cfg.HTTP.Port, defaultHTTPPort)
	}
	if cfg.Postgres.DSN != defaultPostgresDSN {
		t.Fatalf("Postgres.DSN = %q, want %q", cfg.Postgres.DSN, defaultPostgresDSN)
	}
	if cfg.JWT.Secret != defaultJWTSecret {
		t.Fatalf("JWT.Secret = %q, want %q", cfg.JWT.Secret, defaultJWTSecret)
	}
	if cfg.JWT.TTL != time.Duration(defaultJWTTTLMinutes)*time.Minute {
		t.Fatalf("JWT.TTL = %v, want %v", cfg.JWT.TTL, time.Duration(defaultJWTTTLMinutes)*time.Minute)
	}
}

func TestNewConfig_UsesEnvValues(t *testing.T) {
	t.Setenv("APP_HTTP_PORT", "9090")
	t.Setenv("APP_POSTGRES_DSN", "postgres://user:pass@localhost:5432/custom?sslmode=disable")
	t.Setenv("APP_JWT_SECRET", "custom-secret")
	t.Setenv("APP_JWT_TTL_MINUTES", "15")

	cfg := NewConfig()

	if cfg.HTTP.Port != 9090 {
		t.Fatalf("HTTP.Port = %d, want 9090", cfg.HTTP.Port)
	}
	if cfg.Postgres.DSN != "postgres://user:pass@localhost:5432/custom?sslmode=disable" {
		t.Fatalf("Postgres.DSN = %q, want custom DSN", cfg.Postgres.DSN)
	}
	if cfg.JWT.Secret != "custom-secret" {
		t.Fatalf("JWT.Secret = %q, want custom-secret", cfg.JWT.Secret)
	}
	if cfg.JWT.TTL != 15*time.Minute {
		t.Fatalf("JWT.TTL = %v, want %v", cfg.JWT.TTL, 15*time.Minute)
	}
}

func TestGetInt_FallsBackOnInvalidValues(t *testing.T) {
	t.Setenv("APP_HTTP_PORT", "bad")
	if got := getInt("APP_HTTP_PORT", 8080); got != 8080 {
		t.Fatalf("getInt() = %d, want 8080 for invalid input", got)
	}

	t.Setenv("APP_HTTP_PORT", "0")
	if got := getInt("APP_HTTP_PORT", 8080); got != 8080 {
		t.Fatalf("getInt() = %d, want 8080 for non-positive input", got)
	}
}
