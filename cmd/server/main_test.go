package main

import (
	"context"
	"strings"
	"testing"
	"time"

	httptransport "github.com/avito-internships/test-backend-1-EdOoO21/internal/infrastructure/http"
	"github.com/avito-internships/test-backend-1-EdOoO21/internal/settings"
)

type noopLogger struct{}

func (noopLogger) Info(string, ...any)  {}
func (noopLogger) Warn(string, ...any)  {}
func (noopLogger) Error(string, ...any) {}

func TestNewHTTPServer_ConfiguresExpectedFields(t *testing.T) {
	server := newHTTPServer(settings.Config{HTTP: settings.HTTPConfig{Port: 8080}}, httptransport.Services{})

	if server.Addr != ":8080" {
		t.Fatalf("server.Addr = %q, want %q", server.Addr, ":8080")
	}
	if server.Handler == nil {
		t.Fatal("expected server handler to be configured")
	}
	if server.ReadHeaderTimeout != 5*time.Second {
		t.Fatalf("server.ReadHeaderTimeout = %v, want %v", server.ReadHeaderTimeout, 5*time.Second)
	}
}

func TestBuildServices_ReturnsErrorForInvalidDSN(t *testing.T) {
	t.Parallel()

	_, _, err := buildServices(context.Background(), settings.Config{
		Postgres: settings.PostgresConfig{DSN: "://bad dsn"},
		JWT:      settings.JWTConfig{Secret: "test-secret", TTL: time.Hour},
	}, noopLogger{})
	if err == nil {
		t.Fatal("expected error for invalid DSN")
	}
	if !strings.Contains(err.Error(), "connect postgres") {
		t.Fatalf("error = %q, want connect postgres prefix", err)
	}
}
