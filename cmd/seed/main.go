package main

import (
	"context"
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

func main() {
	ctx := context.Background()
	cfg := settings.NewConfig()

	if strings.TrimSpace(cfg.Postgres.DSN) == "" {
		log.Fatal("APP_POSTGRES_DSN is required")
	}

	db, err := apppostgres.New(ctx, cfg.Postgres.DSN)
	if err != nil {
		log.Fatalf("connect postgres: %v", err)
	}
	defer db.Close()

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

	authService := appauth.NewService(userRepo, ids, clock, passwords, tokens)
	roomsService := approoms.NewService(roomRepo, ids, clock)
	schedulesService := appschedules.NewService(roomRepo, scheduleRepo, slotRepo, txManager, ids, clock)
	slotsService := appslots.NewService(roomRepo, scheduleRepo, slotRepo, txManager, ids, clock)
	bookingsService := appbookings.NewService(bookingRepo, slotRepo, txManager, ids, clock, conferenceLinks)

	admin := shared.Actor{UserID: appauth.DummyAdminUserID, Role: domain.RoleAdmin}

	demoUser, createdUser, err := ensureDemoUser(ctx, authService, userRepo, "demo.user@example.com")
	if err != nil {
		log.Fatalf("ensure demo user: %v", err)
	}

	secondUser, createdSecondUser, err := ensureDemoUser(ctx, authService, userRepo, "demo.user2@example.com")
	if err != nil {
		log.Fatalf("ensure second demo user: %v", err)
	}

	roomSpecs := []seedRoomSpec{
		{
			Name:        "Atlas",
			Description: "Main demo room",
			Capacity:    8,
			Days:        []domain.DayOfWeek{domain.Monday, domain.Tuesday, domain.Wednesday, domain.Thursday, domain.Friday, domain.Saturday, domain.Sunday},
			StartTime:   "09:00",
			EndTime:     "18:00",
		},
		{
			Name:        "Orion",
			Description: "Focus room for smaller meetings",
			Capacity:    4,
			Days:        []domain.DayOfWeek{domain.Monday, domain.Tuesday, domain.Wednesday, domain.Thursday, domain.Friday},
			StartTime:   "10:00",
			EndTime:     "17:00",
		},
	}

	roomsByName := make(map[string]domain.Room, len(roomSpecs))
	createdRooms := 0
	createdSchedules := 0

	for _, spec := range roomSpecs {
		room, created, err := ensureRoom(ctx, roomsService, roomRepo, admin, spec)
		if err != nil {
			log.Fatalf("ensure room %q: %v", spec.Name, err)
		}
		if created {
			createdRooms++
		}
		roomsByName[spec.Name] = room

		createdSchedule, err := ensureSchedule(ctx, schedulesService, scheduleRepo, admin, room.ID, spec)
		if err != nil {
			log.Fatalf("ensure schedule for %q: %v", spec.Name, err)
		}
		if createdSchedule {
			createdSchedules++
		}
	}

	createdBooking, err := ensureDemoBooking(ctx, slotsService, bookingsService, demoUser, roomsByName["Atlas"])
	if err != nil {
		log.Fatalf("ensure demo booking: %v", err)
	}

	fmt.Fprintln(os.Stdout, "seed complete")
	fmt.Fprintf(os.Stdout, "demo_user_email=%s created=%t\n", demoUser.Email, createdUser)
	fmt.Fprintf(os.Stdout, "second_demo_user_email=%s created=%t\n", secondUser.Email, createdSecondUser)
	fmt.Fprintf(os.Stdout, "rooms_created=%d schedules_created=%d booking_created=%t\n", createdRooms, createdSchedules, createdBooking)
	fmt.Fprintf(os.Stdout, "demo_room=%s\n", roomsByName["Atlas"].Name)
}

func ensureDemoUser(ctx context.Context, authService authRegistrar, userRepo userByEmailGetter, email string) (domain.User, bool, error) {
	const password = "demo-pass-123"

	user, _, exists, err := userRepo.GetByEmail(ctx, email)
	if err != nil {
		return domain.User{}, false, err
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
		return domain.User{}, false, err
	}

	return out.User, true, nil
}

func ensureRoom(ctx context.Context, roomsService roomCreator, roomRepo roomLister, admin shared.Actor, spec seedRoomSpec) (domain.Room, bool, error) {
	rooms, err := roomRepo.List(ctx)
	if err != nil {
		return domain.Room{}, false, err
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
		return domain.Room{}, false, err
	}

	return out.Room, true, nil
}

func ensureSchedule(ctx context.Context, schedulesService scheduleCreator, scheduleRepo scheduleByRoomGetter, admin shared.Actor, roomID uuid.UUID, spec seedRoomSpec) (bool, error) {
	_, exists, err := scheduleRepo.GetByRoomID(ctx, roomID)
	if err != nil {
		return false, err
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
		return false, err
	}

	return true, nil
}

func ensureDemoBooking(ctx context.Context, slotsService slotAvailabilityLister, bookingsService bookingService, user domain.User, room domain.Room) (bool, error) {
	actor := shared.Actor{UserID: user.ID, Role: user.Role}

	mine, err := bookingsService.ListMine(ctx, appbookings.ListMineInput{Actor: actor})
	if err != nil {
		return false, err
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
		return false, err
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
		return false, err
	}

	return true, nil
}
