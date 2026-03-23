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

	appauth "github.com/avito-internships/test-backend-1-EdOoO21/internal/application/auth"
	appbookings "github.com/avito-internships/test-backend-1-EdOoO21/internal/application/bookings"
	approoms "github.com/avito-internships/test-backend-1-EdOoO21/internal/application/rooms"
	appschedules "github.com/avito-internships/test-backend-1-EdOoO21/internal/application/schedules"
	appslots "github.com/avito-internships/test-backend-1-EdOoO21/internal/application/slots"
	appclock "github.com/avito-internships/test-backend-1-EdOoO21/internal/infrastructure/clock"
	appconference "github.com/avito-internships/test-backend-1-EdOoO21/internal/infrastructure/conference"
	httptransport "github.com/avito-internships/test-backend-1-EdOoO21/internal/infrastructure/http"
	appid "github.com/avito-internships/test-backend-1-EdOoO21/internal/infrastructure/id"
	appjwt "github.com/avito-internships/test-backend-1-EdOoO21/internal/infrastructure/jwt"
	logs "github.com/avito-internships/test-backend-1-EdOoO21/internal/infrastructure/logger"
	apppostgres "github.com/avito-internships/test-backend-1-EdOoO21/internal/infrastructure/postgres"
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

	services, db, err := buildServices(ctx, cfg, logger)
	if err != nil {
		return err
	}
	defer db.Close()

	server := newHTTPServer(cfg, services)

	logger.Info("http server configured", "port", cfg.HTTP.Port)
	logger.Info("postgres configured", "dsn", cfg.Postgres.DSN)

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

func buildServices(ctx context.Context, cfg settings.Config, logger ports.Logger) (httptransport.Services, *apppostgres.DB, error) {
	db, err := apppostgres.New(ctx, cfg.Postgres.DSN)
	if err != nil {
		return httptransport.Services{}, nil, fmt.Errorf("connect postgres: %w", err)
	}

	clock := appclock.New()
	ids := appid.New()
	tokens := appjwt.New(cfg.JWT.Secret, cfg.JWT.TTL)
	conferenceLinks := appconference.NewMock()
	txManager := apppostgres.NewTxManager(db)

	userRepo := apppostgres.NewUserRepository(db)
	_ = userRepo
	roomRepo := apppostgres.NewRoomRepository(db)
	scheduleRepo := apppostgres.NewScheduleRepository(db)
	slotRepo := apppostgres.NewSlotRepository(db)
	bookingRepo := apppostgres.NewBookingRepository(db)

	services := httptransport.Services{
		Logger:    logger,
		Auth:      appauth.NewService(tokens),
		Rooms:     approoms.NewService(roomRepo, ids, clock),
		Schedules: appschedules.NewService(roomRepo, scheduleRepo, slotRepo, txManager, ids, clock),
		Slots:     appslots.NewService(roomRepo, scheduleRepo, slotRepo, txManager, ids, clock),
		Bookings:  appbookings.NewService(bookingRepo, slotRepo, txManager, ids, clock, conferenceLinks),
	}

	return services, db, nil
}

func newHTTPServer(cfg settings.Config, services httptransport.Services) *stdhttp.Server {
	router := httptransport.NewRouter(services)

	return &stdhttp.Server{
		Addr:              fmt.Sprintf(":%d", cfg.HTTP.Port),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}
}
