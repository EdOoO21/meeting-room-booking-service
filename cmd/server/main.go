package main

import (
	"context"
	"errors"
	"fmt"
	stdhttp "net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httptransport "github.com/avito-internships/test-backend-1-EdOoO21/internal/infrastructure/http"
	logs "github.com/avito-internships/test-backend-1-EdOoO21/internal/infrastructure/logger"
	"github.com/avito-internships/test-backend-1-EdOoO21/internal/ports"
	"github.com/avito-internships/test-backend-1-EdOoO21/internal/settings"
)

func main() {
	logger := logs.NewLogger()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, logger); err != nil {
		logger.Error("application stopped with error", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, logger ports.Logger) error {
	cfg := settings.NewConfig()
	server := newHTTPServer(cfg)

	logger.Info("http server configured", "port", cfg.HTTP.Port)

	serverErrCh := make(chan error, 1)

	go func() {
		logger.Info("http server starting", "addr", server.Addr)

		if err := server.ListenAndServe(); err != nil && !errors.Is(err, stdhttp.ErrServerClosed) {
			serverErrCh <- err
			return
		}

		serverErrCh <- nil
	}()

	select {
	case <-ctx.Done():
		logger.Warn("shutdown signal received")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown server: %w", err)
		}

		logger.Info("http server stopped gracefully")
		return nil
	case err := <-serverErrCh:
		if err != nil {
			return fmt.Errorf("listen server: %w", err)
		}

		logger.Info("http server stopped")
		return nil
	}
}

func newHTTPServer(cfg settings.Config) *stdhttp.Server {
	router := httptransport.NewRouter()

	return &stdhttp.Server{
		Addr:              fmt.Sprintf(":%d", cfg.HTTP.Port),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}
}
