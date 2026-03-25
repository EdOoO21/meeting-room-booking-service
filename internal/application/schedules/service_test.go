package schedules

import (
	"context"
	"errors"
	"testing"
	"time"

	appports "github.com/avito-internships/test-backend-1-EdOoO21/internal/application/ports"
	"github.com/avito-internships/test-backend-1-EdOoO21/internal/application/shared"
	"github.com/avito-internships/test-backend-1-EdOoO21/internal/domain"
	"github.com/google/uuid"
)

type fakeRoomRepository struct {
	createFn  func(ctx context.Context, room domain.Room) error
	getByIDFn func(ctx context.Context, id uuid.UUID) (domain.Room, bool, error)
	listFn    func(ctx context.Context) ([]domain.Room, error)
}

func (f fakeRoomRepository) Create(ctx context.Context, room domain.Room) error {
	if f.createFn != nil {
		return f.createFn(ctx, room)
	}
	return nil
}

func (f fakeRoomRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.Room, bool, error) {
	if f.getByIDFn != nil {
		return f.getByIDFn(ctx, id)
	}
	return domain.Room{}, false, nil
}

func (f fakeRoomRepository) List(ctx context.Context) ([]domain.Room, error) {
	if f.listFn != nil {
		return f.listFn(ctx)
	}
	return nil, nil
}

type fakeScheduleRepository struct {
	createFn      func(ctx context.Context, schedule domain.Schedule) error
	getByRoomIDFn func(ctx context.Context, roomID uuid.UUID) (domain.Schedule, bool, error)
}

func (f fakeScheduleRepository) Create(ctx context.Context, schedule domain.Schedule) error {
	if f.createFn != nil {
		return f.createFn(ctx, schedule)
	}
	return nil
}

func (f fakeScheduleRepository) GetByRoomID(ctx context.Context, roomID uuid.UUID) (domain.Schedule, bool, error) {
	if f.getByRoomIDFn != nil {
		return f.getByRoomIDFn(ctx, roomID)
	}
	return domain.Schedule{}, false, nil
}

type fakeSlotRepository struct {
	createManyFn          func(ctx context.Context, slots []domain.Slot) error
	getByIDFn             func(ctx context.Context, id uuid.UUID) (domain.Slot, bool, error)
	hasAnyByRoomAndDateFn func(ctx context.Context, roomID uuid.UUID, date time.Time) (bool, error)
	listAvailableByDateFn func(ctx context.Context, roomID uuid.UUID, date time.Time) ([]domain.Slot, error)
}

func (f fakeSlotRepository) CreateMany(ctx context.Context, slots []domain.Slot) error {
	if f.createManyFn != nil {
		return f.createManyFn(ctx, slots)
	}
	return nil
}

func (f fakeSlotRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.Slot, bool, error) {
	if f.getByIDFn != nil {
		return f.getByIDFn(ctx, id)
	}
	return domain.Slot{}, false, nil
}

func (f fakeSlotRepository) HasAnyByRoomAndDate(ctx context.Context, roomID uuid.UUID, date time.Time) (bool, error) {
	if f.hasAnyByRoomAndDateFn != nil {
		return f.hasAnyByRoomAndDateFn(ctx, roomID, date)
	}
	return false, nil
}

func (f fakeSlotRepository) ListAvailableByRoomAndDate(ctx context.Context, roomID uuid.UUID, date time.Time) ([]domain.Slot, error) {
	if f.listAvailableByDateFn != nil {
		return f.listAvailableByDateFn(ctx, roomID, date)
	}
	return nil, nil
}

type fakeTxManager struct {
	withinTransactionFn func(ctx context.Context, fn func(ctx context.Context) error) error
}

func (f fakeTxManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	if f.withinTransactionFn != nil {
		return f.withinTransactionFn(ctx, fn)
	}
	return fn(ctx)
}

type fakeIDGenerator struct {
	next  []uuid.UUID
	index int
}

func (f *fakeIDGenerator) NewUUID() uuid.UUID {
	if f.index >= len(f.next) {
		return uuid.New()
	}
	id := f.next[f.index]
	f.index++
	return id
}

type fakeClock struct{ now time.Time }

func (f fakeClock) NowUTC() time.Time { return f.now }

func TestService_Create_Success(t *testing.T) {
	t.Parallel()

	roomID := uuid.New()
	scheduleID := uuid.New()
	slotIDs := []uuid.UUID{}
	for range 14 {
		slotIDs = append(slotIDs, uuid.New())
	}
	ids := &fakeIDGenerator{next: append([]uuid.UUID{scheduleID}, slotIDs...)}
	var createdSchedule domain.Schedule
	var createdSlots []domain.Slot

	service := NewService(
		fakeRoomRepository{getByIDFn: func(_ context.Context, _ uuid.UUID) (domain.Room, bool, error) {
			return domain.Room{ID: roomID}, true, nil
		}},
		fakeScheduleRepository{
			getByRoomIDFn: func(_ context.Context, _ uuid.UUID) (domain.Schedule, bool, error) {
				return domain.Schedule{}, false, nil
			},
			createFn: func(_ context.Context, schedule domain.Schedule) error {
				createdSchedule = schedule
				return nil
			},
		},
		fakeSlotRepository{createManyFn: func(_ context.Context, slots []domain.Slot) error {
			createdSlots = append([]domain.Slot(nil), slots...)
			return nil
		}},
		fakeTxManager{},
		ids,
		fakeClock{now: time.Date(2026, time.March, 23, 8, 0, 0, 0, time.UTC)},
	)

	out, err := service.Create(context.Background(), CreateInput{
		Actor:      shared.Actor{UserID: uuid.New(), Role: domain.RoleAdmin},
		RoomID:     roomID,
		DaysOfWeek: []domain.DayOfWeek{domain.Monday, domain.Tuesday, domain.Wednesday, domain.Thursday, domain.Friday, domain.Saturday, domain.Sunday},
		StartTime:  "09:00",
		EndTime:    "10:00",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if out.Schedule.ID != scheduleID {
		t.Fatalf("out.Schedule.ID = %v, want %v", out.Schedule.ID, scheduleID)
	}

	if out.GeneratedSlots != 14 {
		t.Fatalf("out.GeneratedSlots = %d, want %d", out.GeneratedSlots, 14)
	}

	if createdSchedule.ID != scheduleID {
		t.Fatalf("createdSchedule.ID = %v, want %v", createdSchedule.ID, scheduleID)
	}

	if len(createdSlots) != 14 {
		t.Fatalf("len(createdSlots) = %d, want %d", len(createdSlots), 14)
	}
}

func TestService_Create_ReturnsRoomNotFoundAndScheduleExists(t *testing.T) {
	t.Parallel()

	t.Run("room not found", func(t *testing.T) {
		t.Parallel()

		service := NewService(
			fakeRoomRepository{getByIDFn: func(_ context.Context, _ uuid.UUID) (domain.Room, bool, error) {
				return domain.Room{}, false, nil
			}},
			fakeScheduleRepository{},
			fakeSlotRepository{},
			fakeTxManager{},
			&fakeIDGenerator{next: []uuid.UUID{uuid.New()}},
			fakeClock{now: time.Date(2026, time.March, 23, 8, 0, 0, 0, time.UTC)},
		)

		_, err := service.Create(context.Background(), CreateInput{
			Actor:      shared.Actor{UserID: uuid.New(), Role: domain.RoleAdmin},
			RoomID:     uuid.New(),
			DaysOfWeek: []domain.DayOfWeek{domain.Monday},
			StartTime:  "09:00",
			EndTime:    "10:00",
		})
		if !errors.Is(err, shared.ErrRoomNotFound) {
			t.Fatalf("Create() error = %v, want %v", err, shared.ErrRoomNotFound)
		}
	})

	t.Run("schedule exists", func(t *testing.T) {
		t.Parallel()

		service := NewService(
			fakeRoomRepository{getByIDFn: func(_ context.Context, id uuid.UUID) (domain.Room, bool, error) {
				return domain.Room{ID: id}, true, nil
			}},
			fakeScheduleRepository{getByRoomIDFn: func(_ context.Context, _ uuid.UUID) (domain.Schedule, bool, error) {
				return domain.Schedule{ID: uuid.New()}, true, nil
			}},
			fakeSlotRepository{},
			fakeTxManager{},
			&fakeIDGenerator{next: []uuid.UUID{uuid.New()}},
			fakeClock{now: time.Date(2026, time.March, 23, 8, 0, 0, 0, time.UTC)},
		)

		_, err := service.Create(context.Background(), CreateInput{
			Actor:      shared.Actor{UserID: uuid.New(), Role: domain.RoleAdmin},
			RoomID:     uuid.New(),
			DaysOfWeek: []domain.DayOfWeek{domain.Monday},
			StartTime:  "09:00",
			EndTime:    "10:00",
		})
		if !errors.Is(err, shared.ErrScheduleExists) {
			t.Fatalf("Create() error = %v, want %v", err, shared.ErrScheduleExists)
		}
	})
}

func TestGenerateSlotsForDate(t *testing.T) {
	t.Parallel()

	schedule, err := domain.NewSchedule(
		uuid.New(),
		uuid.New(),
		[]domain.DayOfWeek{domain.Monday},
		domain.TimeOfDay{Hour: 9, Minute: 0},
		domain.TimeOfDay{Hour: 10, Minute: 0},
	)
	if err != nil {
		t.Fatalf("NewSchedule() error = %v", err)
	}

	ids := &fakeIDGenerator{next: []uuid.UUID{uuid.New(), uuid.New(), uuid.New(), uuid.New()}}
	date := time.Date(2026, time.March, 23, 0, 0, 0, 0, time.UTC)

	slots, err := GenerateSlotsForDate(ids, schedule, date)
	if err != nil {
		t.Fatalf("GenerateSlotsForDate() error = %v", err)
	}

	if len(slots) != 2 {
		t.Fatalf("len(slots) = %d, want %d", len(slots), 2)
	}

	empty, err := GenerateSlotsForDate(ids, schedule, date.AddDate(0, 0, 1))
	if err != nil {
		t.Fatalf("GenerateSlotsForDate() on non matching date error = %v", err)
	}

	if len(empty) != 0 {
		t.Fatalf("len(empty) = %d, want 0", len(empty))
	}
}

var _ appports.RoomRepository = fakeRoomRepository{}
var _ appports.ScheduleRepository = fakeScheduleRepository{}
var _ appports.SlotRepository = fakeSlotRepository{}
var _ appports.TxManager = fakeTxManager{}
var _ appports.Clock = fakeClock{}
