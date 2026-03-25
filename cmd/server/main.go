// @title Room Booking API
// @version 1.0
// @description API для бронирования переговорок и управления расписаниями.
// @host localhost:8080
// @BasePath /
// @schemes http
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
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
	apppassword "github.com/avito-internships/test-backend-1-EdOoO21/internal/infrastructure/password"
	apppostgres "github.com/avito-internships/test-backend-1-EdOoO21/internal/infrastructure/postgres"
	"github.com/avito-internships/test-backend-1-EdOoO21/internal/ports"
	"github.com/avito-internships/test-backend-1-EdOoO21/internal/settings"
)

const (
	serverShutdownTimeout   = 10 * time.Second
	serverReadHeaderTimeout = 5 * time.Second
)

func main() {
	logger := logs.NewLogger()

	if err := run(logger); err != nil {
		logger.Error("application stopped with error", "error", err)
		os.Exit(1)
	}
}

func run(logger ports.Logger) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

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

		if listenErr := server.ListenAndServe(); listenErr != nil && !errors.Is(listenErr, stdhttp.ErrServerClosed) {
			serverErrCh <- listenErr
			return
		}

		serverErrCh <- nil
	}()

	select {
	case <-ctx.Done():
		logger.Warn("shutdown signal received")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), serverShutdownTimeout)
		defer cancel()

		if shutdownErr := server.Shutdown(shutdownCtx); shutdownErr != nil {
			return fmt.Errorf("shutdown server: %w", shutdownErr)
		}

		logger.Info("http server stopped gracefully")
		return nil
	case serverErr := <-serverErrCh:
		if serverErr != nil {
			return fmt.Errorf("listen server: %w", serverErr)
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
	passwords := apppassword.New()
	tokens := appjwt.New(cfg.JWT.Secret, cfg.JWT.TTL)
	conferenceLinks := appconference.NewMock()
	txManager := apppostgres.NewTxManager(db)

	userRepo := apppostgres.NewUserRepository(db)
	roomRepo := apppostgres.NewRoomRepository(db)
	scheduleRepo := apppostgres.NewScheduleRepository(db)
	slotRepo := apppostgres.NewSlotRepository(db)
	bookingRepo := apppostgres.NewBookingRepository(db)

	services := httptransport.Services{
		Logger:    logger,
		JWT:       tokens,
		Auth:      appauth.NewService(userRepo, ids, clock, passwords, tokens),
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
		ReadHeaderTimeout: serverReadHeaderTimeout,
	}
}
