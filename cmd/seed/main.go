package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	appauth "github.com/avito-internships/test-backend-1-EdOoO21/internal/application/auth"
	appbookings "github.com/avito-internships/test-backend-1-EdOoO21/internal/application/bookings"
	approoms "github.com/avito-internships/test-backend-1-EdOoO21/internal/application/rooms"
	appschedules "github.com/avito-internships/test-backend-1-EdOoO21/internal/application/schedules"
	"github.com/avito-internships/test-backend-1-EdOoO21/internal/application/shared"
	appslots "github.com/avito-internships/test-backend-1-EdOoO21/internal/application/slots"
	"github.com/avito-internships/test-backend-1-EdOoO21/internal/domain"
	appclock "github.com/avito-internships/test-backend-1-EdOoO21/internal/infrastructure/clock"
	appconference "github.com/avito-internships/test-backend-1-EdOoO21/internal/infrastructure/conference"
	appid "github.com/avito-internships/test-backend-1-EdOoO21/internal/infrastructure/id"
	appjwt "github.com/avito-internships/test-backend-1-EdOoO21/internal/infrastructure/jwt"
	apppassword "github.com/avito-internships/test-backend-1-EdOoO21/internal/infrastructure/password"
	apppostgres "github.com/avito-internships/test-backend-1-EdOoO21/internal/infrastructure/postgres"
	"github.com/avito-internships/test-backend-1-EdOoO21/internal/settings"
	"github.com/google/uuid"
)

const (
	atlasRoomCapacity = 8
	orionRoomCapacity = 4
)

type seedRoomSpec struct {
	Name        string
	Description string
	Capacity    int
	Days        []domain.DayOfWeek
	StartTime   string
	EndTime     string
}

type authRegistrar interface {
	Register(ctx context.Context, input appauth.RegisterInput) (appauth.RegisterOutput, error)
}

type userByEmailGetter interface {
	GetByEmail(ctx context.Context, email string) (domain.User, string, bool, error)
}

type roomCreator interface {
	Create(ctx context.Context, input approoms.CreateInput) (approoms.CreateOutput, error)
}

type roomLister interface {
	List(ctx context.Context) ([]domain.Room, error)
}

type scheduleCreator interface {
	Create(ctx context.Context, input appschedules.CreateInput) (appschedules.CreateOutput, error)
}

type scheduleByRoomGetter interface {
	GetByRoomID(ctx context.Context, roomID uuid.UUID) (domain.Schedule, bool, error)
}

type slotAvailabilityLister interface {
	ListAvailable(ctx context.Context, input appslots.ListAvailableInput) (appslots.ListAvailableOutput, error)
}

type bookingService interface {
	ListMine(ctx context.Context, input appbookings.ListMineInput) (appbookings.ListMineOutput, error)
	Create(ctx context.Context, input appbookings.CreateInput) (appbookings.CreateOutput, error)
}

type seedDependencies struct {
	authService      authRegistrar
	userRepo         userByEmailGetter
	roomsService     roomCreator
	roomRepo         roomLister
	schedulesService scheduleCreator
	scheduleRepo     scheduleByRoomGetter
	slotsService     slotAvailabilityLister
	bookingsService  bookingService
	admin            shared.Actor
}

type seedResult struct {
	demoUser           domain.User
	createdUser        bool
	secondUser         domain.User
	createdSecondUser  bool
	roomsByName        map[string]domain.Room
	createdRooms       int
	createdSchedules   int
	createdDemoBooking bool
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx := context.Background()
	cfg := settings.NewConfig()

	if strings.TrimSpace(cfg.Postgres.DSN) == "" {
		return errors.New("APP_POSTGRES_DSN is required")
	}

	db, err := apppostgres.New(ctx, cfg.Postgres.DSN)
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer db.Close()

	deps := buildSeedDependencies(cfg, db)
	result, err := seedDemoData(ctx, deps)
	if err != nil {
		return err
	}

	writeSeedSummary(result)
	return nil
}

func buildSeedDependencies(cfg settings.Config, db *apppostgres.DB) seedDependencies {
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

	return seedDependencies{
		authService:      appauth.NewService(userRepo, ids, clock, passwords, tokens),
		userRepo:         userRepo,
		roomsService:     approoms.NewService(roomRepo, ids, clock),
		roomRepo:         roomRepo,
		schedulesService: appschedules.NewService(roomRepo, scheduleRepo, slotRepo, txManager, ids, clock),
		scheduleRepo:     scheduleRepo,
		slotsService:     appslots.NewService(roomRepo, scheduleRepo, slotRepo, txManager, ids, clock),
		bookingsService:  appbookings.NewService(bookingRepo, slotRepo, txManager, ids, clock, conferenceLinks),
		admin:            shared.Actor{UserID: appauth.DummyAdminUserID, Role: domain.RoleAdmin},
	}
}

func seedDemoData(ctx context.Context, deps seedDependencies) (seedResult, error) {
	result := seedResult{}

	var err error
	result.demoUser, result.createdUser, err = ensureDemoUser(ctx, deps.authService, deps.userRepo, "demo.user@example.com")
	if err != nil {
		return seedResult{}, fmt.Errorf("ensure demo user: %w", err)
	}

	result.secondUser, result.createdSecondUser, err = ensureDemoUser(ctx, deps.authService, deps.userRepo, "demo.user2@example.com")
	if err != nil {
		return seedResult{}, fmt.Errorf("ensure second demo user: %w", err)
	}

	result.roomsByName, result.createdRooms, result.createdSchedules, err = seedRooms(ctx, deps)
	if err != nil {
		return seedResult{}, err
	}

	result.createdDemoBooking, err = ensureDemoBooking(ctx, deps.slotsService, deps.bookingsService, result.demoUser, result.roomsByName["Atlas"])
	if err != nil {
		return seedResult{}, fmt.Errorf("ensure demo booking: %w", err)
	}

	return result, nil
}

func seedRooms(ctx context.Context, deps seedDependencies) (map[string]domain.Room, int, int, error) {
	roomsByName := make(map[string]domain.Room, len(defaultSeedRoomSpecs()))
	createdRooms := 0
	createdSchedules := 0

	for _, spec := range defaultSeedRoomSpecs() {
		room, created, roomErr := ensureRoom(ctx, deps.roomsService, deps.roomRepo, deps.admin, spec)
		if roomErr != nil {
			return nil, 0, 0, fmt.Errorf("ensure room %q: %w", spec.Name, roomErr)
		}
		if created {
			createdRooms++
		}
		roomsByName[spec.Name] = room

		createdSchedule, scheduleErr := ensureSchedule(ctx, deps.schedulesService, deps.scheduleRepo, deps.admin, room.ID, spec)
		if scheduleErr != nil {
			return nil, 0, 0, fmt.Errorf("ensure schedule for %q: %w", spec.Name, scheduleErr)
		}
		if createdSchedule {
			createdSchedules++
		}
	}

	return roomsByName, createdRooms, createdSchedules, nil
}

func defaultSeedRoomSpecs() []seedRoomSpec {
	return []seedRoomSpec{
		{
			Name:        "Atlas",
			Description: "Main demo room",
			Capacity:    atlasRoomCapacity,
			Days:        []domain.DayOfWeek{domain.Monday, domain.Tuesday, domain.Wednesday, domain.Thursday, domain.Friday, domain.Saturday, domain.Sunday},
			StartTime:   "09:00",
			EndTime:     "18:00",
		},
		{
			Name:        "Orion",
			Description: "Focus room for smaller meetings",
			Capacity:    orionRoomCapacity,
			Days:        []domain.DayOfWeek{domain.Monday, domain.Tuesday, domain.Wednesday, domain.Thursday, domain.Friday},
			StartTime:   "10:00",
			EndTime:     "17:00",
		},
	}
}

func writeSeedSummary(result seedResult) {
	mustWriteSeedLinef("seed complete\n")
	mustWriteSeedLinef("demo_user_email=%s created=%t\n", result.demoUser.Email, result.createdUser)
	mustWriteSeedLinef("second_demo_user_email=%s created=%t\n", result.secondUser.Email, result.createdSecondUser)
	mustWriteSeedLinef("rooms_created=%d schedules_created=%d booking_created=%t\n", result.createdRooms, result.createdSchedules, result.createdDemoBooking)
	mustWriteSeedLinef("demo_room=%s\n", result.roomsByName["Atlas"].Name)
}

func mustWriteSeedLinef(format string, args ...any) {
	if _, err := fmt.Fprintf(os.Stdout, format, args...); err != nil {
		log.Printf("write seed output: %v", err)
		os.Exit(1)
	}
}

func ensureDemoUser(ctx context.Context, authService authRegistrar, userRepo userByEmailGetter, email string) (domain.User, bool, error) {
	const password = "demo-pass-123"

	user, _, exists, err := userRepo.GetByEmail(ctx, email)
	if err != nil {
		return domain.User{}, false, fmt.Errorf("get user by email: %w", err)
	}
	if exists {
		return user, false, nil
	}

	out, err := authService.Register(ctx, appauth.RegisterInput{
		Email:    email,
		Password: password,
		Role:     domain.RoleUser,
	})
	if err != nil {
		return domain.User{}, false, fmt.Errorf("register demo user: %w", err)
	}

	return out.User, true, nil
}

func ensureRoom(ctx context.Context, roomsService roomCreator, roomRepo roomLister, admin shared.Actor, spec seedRoomSpec) (domain.Room, bool, error) {
	rooms, err := roomRepo.List(ctx)
	if err != nil {
		return domain.Room{}, false, fmt.Errorf("list rooms: %w", err)
	}

	for _, room := range rooms {
		if strings.EqualFold(room.Name, spec.Name) {
			return room, false, nil
		}
	}

	capacity := spec.Capacity
	out, err := roomsService.Create(ctx, approoms.CreateInput{
		Actor:       admin,
		Name:        spec.Name,
		Description: spec.Description,
		Capacity:    &capacity,
	})
	if err != nil {
		return domain.Room{}, false, fmt.Errorf("create room: %w", err)
	}

	return out.Room, true, nil
}

func ensureSchedule(ctx context.Context, schedulesService scheduleCreator, scheduleRepo scheduleByRoomGetter, admin shared.Actor, roomID uuid.UUID, spec seedRoomSpec) (bool, error) {
	_, exists, err := scheduleRepo.GetByRoomID(ctx, roomID)
	if err != nil {
		return false, fmt.Errorf("get schedule by room: %w", err)
	}
	if exists {
		return false, nil
	}

	_, err = schedulesService.Create(ctx, appschedules.CreateInput{
		Actor:      admin,
		RoomID:     roomID,
		DaysOfWeek: spec.Days,
		StartTime:  spec.StartTime,
		EndTime:    spec.EndTime,
	})
	if err != nil {
		return false, fmt.Errorf("create schedule: %w", err)
	}

	return true, nil
}

func ensureDemoBooking(ctx context.Context, slotsService slotAvailabilityLister, bookingsService bookingService, user domain.User, room domain.Room) (bool, error) {
	actor := shared.Actor{UserID: user.ID, Role: user.Role}

	mine, err := bookingsService.ListMine(ctx, appbookings.ListMineInput{Actor: actor})
	if err != nil {
		return false, fmt.Errorf("list demo bookings: %w", err)
	}
	if len(mine.Bookings) > 0 {
		return false, nil
	}

	date := time.Now().UTC().AddDate(0, 0, 1)
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)

	available, err := slotsService.ListAvailable(ctx, appslots.ListAvailableInput{
		Actor:  actor,
		RoomID: room.ID,
		Date:   date,
	})
	if err != nil {
		return false, fmt.Errorf("list available slots: %w", err)
	}
	if len(available.Slots) == 0 {
		return false, nil
	}

	_, err = bookingsService.Create(ctx, appbookings.CreateInput{
		Actor:                actor,
		SlotID:               available.Slots[0].ID,
		CreateConferenceLink: true,
	})
	if err != nil {
		return false, fmt.Errorf("create demo booking: %w", err)
	}

	return true, nil
}
