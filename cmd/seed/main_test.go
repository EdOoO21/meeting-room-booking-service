package main

import (
	"context"
	"errors"
	"testing"
	"time"

	appauth "github.com/avito-internships/test-backend-1-EdOoO21/internal/application/auth"
	appbookings "github.com/avito-internships/test-backend-1-EdOoO21/internal/application/bookings"
	approoms "github.com/avito-internships/test-backend-1-EdOoO21/internal/application/rooms"
	appschedules "github.com/avito-internships/test-backend-1-EdOoO21/internal/application/schedules"
	"github.com/avito-internships/test-backend-1-EdOoO21/internal/application/shared"
	appslots "github.com/avito-internships/test-backend-1-EdOoO21/internal/application/slots"
	"github.com/avito-internships/test-backend-1-EdOoO21/internal/domain"
	"github.com/google/uuid"
)

type fakeAuthRegistrar struct {
	registerFn func(ctx context.Context, input appauth.RegisterInput) (appauth.RegisterOutput, error)
}

func (f fakeAuthRegistrar) Register(ctx context.Context, input appauth.RegisterInput) (appauth.RegisterOutput, error) {
	return f.registerFn(ctx, input)
}

type fakeUserByEmailGetter struct {
	getByEmailFn func(ctx context.Context, email string) (domain.User, string, bool, error)
}

func (f fakeUserByEmailGetter) GetByEmail(ctx context.Context, email string) (domain.User, string, bool, error) {
	return f.getByEmailFn(ctx, email)
}

type fakeRoomCreator struct {
	createFn func(ctx context.Context, input approoms.CreateInput) (approoms.CreateOutput, error)
}

func (f fakeRoomCreator) Create(ctx context.Context, input approoms.CreateInput) (approoms.CreateOutput, error) {
	return f.createFn(ctx, input)
}

type fakeRoomLister struct {
	listFn func(ctx context.Context) ([]domain.Room, error)
}

func (f fakeRoomLister) List(ctx context.Context) ([]domain.Room, error) {
	return f.listFn(ctx)
}

type fakeScheduleCreator struct {
	createFn func(ctx context.Context, input appschedules.CreateInput) (appschedules.CreateOutput, error)
}

func (f fakeScheduleCreator) Create(ctx context.Context, input appschedules.CreateInput) (appschedules.CreateOutput, error) {
	return f.createFn(ctx, input)
}

type fakeScheduleByRoomGetter struct {
	getByRoomIDFn func(ctx context.Context, roomID uuid.UUID) (domain.Schedule, bool, error)
}

func (f fakeScheduleByRoomGetter) GetByRoomID(ctx context.Context, roomID uuid.UUID) (domain.Schedule, bool, error) {
	return f.getByRoomIDFn(ctx, roomID)
}

type fakeSlotAvailabilityLister struct {
	listAvailableFn func(ctx context.Context, input appslots.ListAvailableInput) (appslots.ListAvailableOutput, error)
}

func (f fakeSlotAvailabilityLister) ListAvailable(ctx context.Context, input appslots.ListAvailableInput) (appslots.ListAvailableOutput, error) {
	return f.listAvailableFn(ctx, input)
}

type fakeBookingService struct {
	listMineFn func(ctx context.Context, input appbookings.ListMineInput) (appbookings.ListMineOutput, error)
	createFn   func(ctx context.Context, input appbookings.CreateInput) (appbookings.CreateOutput, error)
}

func (f fakeBookingService) ListMine(ctx context.Context, input appbookings.ListMineInput) (appbookings.ListMineOutput, error) {
	return f.listMineFn(ctx, input)
}

func (f fakeBookingService) Create(ctx context.Context, input appbookings.CreateInput) (appbookings.CreateOutput, error) {
	return f.createFn(ctx, input)
}

func TestEnsureDemoUser_ReturnsExistingUser(t *testing.T) {
	t.Parallel()

	existing, err := domain.NewUser(uuid.New(), "demo.user@example.com", domain.RoleUser, time.Date(2026, time.March, 24, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("NewUser() error = %v", err)
	}

	called := false
	user, created, err := ensureDemoUser(context.Background(), fakeAuthRegistrar{
		registerFn: func(ctx context.Context, input appauth.RegisterInput) (appauth.RegisterOutput, error) {
			called = true
			return appauth.RegisterOutput{}, nil
		},
	}, fakeUserByEmailGetter{
		getByEmailFn: func(ctx context.Context, email string) (domain.User, string, bool, error) {
			if email != "demo.user@example.com" {
				t.Fatalf("email = %q, want %q", email, "demo.user@example.com")
			}
			return existing, "", true, nil
		},
	}, "demo.user@example.com")
	if err != nil {
		t.Fatalf("ensureDemoUser() error = %v", err)
	}
	if created {
		t.Fatal("expected created=false for existing user")
	}
	if called {
		t.Fatal("Register() should not be called for existing user")
	}
	if user.ID != existing.ID {
		t.Fatalf("user.ID = %v, want %v", user.ID, existing.ID)
	}
}

func TestEnsureDemoUser_RegistersMissingUser(t *testing.T) {
	t.Parallel()

	createdUser, err := domain.NewUser(uuid.New(), "demo.user@example.com", domain.RoleUser, time.Date(2026, time.March, 24, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("NewUser() error = %v", err)
	}

	called := false
	user, created, err := ensureDemoUser(context.Background(), fakeAuthRegistrar{
		registerFn: func(ctx context.Context, input appauth.RegisterInput) (appauth.RegisterOutput, error) {
			called = true
			if input.Email != "demo.user@example.com" || input.Password != "demo-pass-123" || input.Role != domain.RoleUser {
				t.Fatalf("unexpected register input: %+v", input)
			}
			return appauth.RegisterOutput{User: createdUser}, nil
		},
	}, fakeUserByEmailGetter{
		getByEmailFn: func(ctx context.Context, email string) (domain.User, string, bool, error) {
			if email != "demo.user@example.com" {
				t.Fatalf("email = %q, want %q", email, "demo.user@example.com")
			}
			return domain.User{}, "", false, nil
		},
	}, "demo.user@example.com")
	if err != nil {
		t.Fatalf("ensureDemoUser() error = %v", err)
	}
	if !created {
		t.Fatal("expected created=true for new user")
	}
	if !called {
		t.Fatal("expected Register() to be called")
	}
	if user.ID != createdUser.ID {
		t.Fatalf("user.ID = %v, want %v", user.ID, createdUser.ID)
	}
}

func TestEnsureRoom_ReturnsExistingRoomByName(t *testing.T) {
	t.Parallel()

	existing := domain.Room{ID: uuid.New(), Name: "Atlas"}
	created := false
	room, wasCreated, err := ensureRoom(context.Background(), fakeRoomCreator{
		createFn: func(ctx context.Context, input approoms.CreateInput) (approoms.CreateOutput, error) {
			created = true
			return approoms.CreateOutput{}, nil
		},
	}, fakeRoomLister{
		listFn: func(ctx context.Context) ([]domain.Room, error) {
			return []domain.Room{existing}, nil
		},
	}, shared.Actor{UserID: uuid.New(), Role: domain.RoleAdmin}, seedRoomSpec{Name: "atlas"})
	if err != nil {
		t.Fatalf("ensureRoom() error = %v", err)
	}
	if wasCreated {
		t.Fatal("expected existing room to be reused")
	}
	if created {
		t.Fatal("Create() should not be called when room already exists")
	}
	if room.ID != existing.ID {
		t.Fatalf("room.ID = %v, want %v", room.ID, existing.ID)
	}
}

func TestEnsureRoom_CreatesMissingRoom(t *testing.T) {
	t.Parallel()

	createdRoom := domain.Room{ID: uuid.New(), Name: "Atlas"}
	room, wasCreated, err := ensureRoom(context.Background(), fakeRoomCreator{
		createFn: func(ctx context.Context, input approoms.CreateInput) (approoms.CreateOutput, error) {
			if input.Name != "Atlas" || input.Description != "Main demo room" || input.Capacity == nil || *input.Capacity != 8 {
				t.Fatalf("unexpected create input: %+v", input)
			}
			return approoms.CreateOutput{Room: createdRoom}, nil
		},
	}, fakeRoomLister{listFn: func(ctx context.Context) ([]domain.Room, error) { return nil, nil }}, shared.Actor{UserID: uuid.New(), Role: domain.RoleAdmin}, seedRoomSpec{Name: "Atlas", Description: "Main demo room", Capacity: 8})
	if err != nil {
		t.Fatalf("ensureRoom() error = %v", err)
	}
	if !wasCreated {
		t.Fatal("expected room to be created")
	}
	if room.ID != createdRoom.ID {
		t.Fatalf("room.ID = %v, want %v", room.ID, createdRoom.ID)
	}
}

func TestEnsureSchedule_SkipsExistingSchedule(t *testing.T) {
	t.Parallel()

	created := false
	createdSchedule, err := ensureSchedule(context.Background(), fakeScheduleCreator{
		createFn: func(ctx context.Context, input appschedules.CreateInput) (appschedules.CreateOutput, error) {
			created = true
			return appschedules.CreateOutput{}, nil
		},
	}, fakeScheduleByRoomGetter{
		getByRoomIDFn: func(ctx context.Context, roomID uuid.UUID) (domain.Schedule, bool, error) {
			return domain.Schedule{ID: uuid.New()}, true, nil
		},
	}, shared.Actor{UserID: uuid.New(), Role: domain.RoleAdmin}, uuid.New(), seedRoomSpec{})
	if err != nil {
		t.Fatalf("ensureSchedule() error = %v", err)
	}
	if createdSchedule {
		t.Fatal("expected existing schedule to be reused")
	}
	if created {
		t.Fatal("Create() should not be called when schedule already exists")
	}
}

func TestEnsureSchedule_CreatesMissingSchedule(t *testing.T) {
	t.Parallel()

	roomID := uuid.New()
	createdSchedule, err := ensureSchedule(context.Background(), fakeScheduleCreator{
		createFn: func(ctx context.Context, input appschedules.CreateInput) (appschedules.CreateOutput, error) {
			if input.RoomID != roomID || input.StartTime != "09:00" || input.EndTime != "18:00" || len(input.DaysOfWeek) != 2 {
				t.Fatalf("unexpected create input: %+v", input)
			}
			return appschedules.CreateOutput{}, nil
		},
	}, fakeScheduleByRoomGetter{
		getByRoomIDFn: func(ctx context.Context, roomID uuid.UUID) (domain.Schedule, bool, error) {
			return domain.Schedule{}, false, nil
		},
	}, shared.Actor{UserID: uuid.New(), Role: domain.RoleAdmin}, roomID, seedRoomSpec{Days: []domain.DayOfWeek{domain.Monday, domain.Tuesday}, StartTime: "09:00", EndTime: "18:00"})
	if err != nil {
		t.Fatalf("ensureSchedule() error = %v", err)
	}
	if !createdSchedule {
		t.Fatal("expected schedule to be created")
	}
}

func TestEnsureDemoBooking_SkipsWhenUserAlreadyHasBookings(t *testing.T) {
	t.Parallel()

	user, err := domain.NewUser(uuid.New(), "demo.user@example.com", domain.RoleUser, time.Date(2026, time.March, 24, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("NewUser() error = %v", err)
	}

	slotsCalled := false
	createCalled := false
	created, err := ensureDemoBooking(context.Background(), fakeSlotAvailabilityLister{
		listAvailableFn: func(ctx context.Context, input appslots.ListAvailableInput) (appslots.ListAvailableOutput, error) {
			slotsCalled = true
			return appslots.ListAvailableOutput{}, nil
		},
	}, fakeBookingService{
		listMineFn: func(ctx context.Context, input appbookings.ListMineInput) (appbookings.ListMineOutput, error) {
			return appbookings.ListMineOutput{Bookings: []domain.Booking{{ID: uuid.New()}}}, nil
		},
		createFn: func(ctx context.Context, input appbookings.CreateInput) (appbookings.CreateOutput, error) {
			createCalled = true
			return appbookings.CreateOutput{}, nil
		},
	}, user, domain.Room{ID: uuid.New()})
	if err != nil {
		t.Fatalf("ensureDemoBooking() error = %v", err)
	}
	if created {
		t.Fatal("expected no booking to be created")
	}
	if slotsCalled || createCalled {
		t.Fatal("slot listing and booking creation should not run when user already has future bookings")
	}
}

func TestEnsureDemoBooking_CreatesBookingFromFirstAvailableSlot(t *testing.T) {
	t.Parallel()

	user, err := domain.NewUser(uuid.New(), "demo.user@example.com", domain.RoleUser, time.Date(2026, time.March, 24, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("NewUser() error = %v", err)
	}

	roomID := uuid.New()
	slotID := uuid.New()
	created, err := ensureDemoBooking(context.Background(), fakeSlotAvailabilityLister{
		listAvailableFn: func(ctx context.Context, input appslots.ListAvailableInput) (appslots.ListAvailableOutput, error) {
			if input.RoomID != roomID {
				t.Fatalf("input.RoomID = %v, want %v", input.RoomID, roomID)
			}
			if input.Actor.UserID != user.ID || input.Actor.Role != domain.RoleUser {
				t.Fatalf("unexpected actor: %+v", input.Actor)
			}
			if input.Date.Location() != time.UTC || input.Date.Hour() != 0 || input.Date.Minute() != 0 {
				t.Fatalf("expected normalized UTC date, got %v", input.Date)
			}
			return appslots.ListAvailableOutput{Slots: []domain.Slot{{ID: slotID}}}, nil
		},
	}, fakeBookingService{
		listMineFn: func(ctx context.Context, input appbookings.ListMineInput) (appbookings.ListMineOutput, error) {
			return appbookings.ListMineOutput{}, nil
		},
		createFn: func(ctx context.Context, input appbookings.CreateInput) (appbookings.CreateOutput, error) {
			if input.SlotID != slotID || !input.CreateConferenceLink {
				t.Fatalf("unexpected create input: %+v", input)
			}
			return appbookings.CreateOutput{}, nil
		},
	}, user, domain.Room{ID: roomID})
	if err != nil {
		t.Fatalf("ensureDemoBooking() error = %v", err)
	}
	if !created {
		t.Fatal("expected booking to be created")
	}
}

func TestEnsureDemoBooking_PropagatesCreateError(t *testing.T) {
	t.Parallel()

	user, err := domain.NewUser(uuid.New(), "demo.user@example.com", domain.RoleUser, time.Date(2026, time.March, 24, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("NewUser() error = %v", err)
	}

	wantErr := errors.New("create booking failed")
	_, err = ensureDemoBooking(context.Background(), fakeSlotAvailabilityLister{
		listAvailableFn: func(ctx context.Context, input appslots.ListAvailableInput) (appslots.ListAvailableOutput, error) {
			return appslots.ListAvailableOutput{Slots: []domain.Slot{{ID: uuid.New()}}}, nil
		},
	}, fakeBookingService{
		listMineFn: func(ctx context.Context, input appbookings.ListMineInput) (appbookings.ListMineOutput, error) {
			return appbookings.ListMineOutput{}, nil
		},
		createFn: func(ctx context.Context, input appbookings.CreateInput) (appbookings.CreateOutput, error) {
			return appbookings.CreateOutput{}, wantErr
		},
	}, user, domain.Room{ID: uuid.New()})
	if !errors.Is(err, wantErr) {
		t.Fatalf("ensureDemoBooking() error = %v, want %v", err, wantErr)
	}
}
