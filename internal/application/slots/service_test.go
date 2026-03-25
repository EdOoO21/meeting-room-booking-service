package slots

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
	listAvailableFn       func(ctx context.Context, roomID uuid.UUID, date time.Time) ([]domain.Slot, error)
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
	if f.listAvailableFn != nil {
		return f.listAvailableFn(ctx, roomID, date)
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
		return uuid.Nil
	}
	id := f.next[f.index]
	f.index++
	return id
}

type fakeClock struct{ now time.Time }

func (f fakeClock) NowUTC() time.Time { return f.now }

func TestService_ListAvailable_ReturnsExistingSlots(t *testing.T) {
	t.Parallel()

	roomID := uuid.New()
	date := time.Date(2026, time.March, 24, 15, 0, 0, 0, time.UTC)
	wantSlots := []domain.Slot{{ID: uuid.New()}, {ID: uuid.New()}}
	createManyCalled := false

	service := NewService(
		fakeRoomRepository{getByIDFn: func(_ context.Context, _ uuid.UUID) (domain.Room, bool, error) {
			return domain.Room{ID: roomID}, true, nil
		}},
		fakeScheduleRepository{},
		fakeSlotRepository{
			hasAnyByRoomAndDateFn: func(_ context.Context, _ uuid.UUID, _ time.Time) (bool, error) {
				return true, nil
			},
			createManyFn: func(_ context.Context, _ []domain.Slot) error {
				createManyCalled = true
				return nil
			},
			listAvailableFn: func(_ context.Context, _ uuid.UUID, _ time.Time) ([]domain.Slot, error) {
				return wantSlots, nil
			},
		},
		fakeTxManager{},
		&fakeIDGenerator{},
		fakeClock{now: time.Date(2026, time.March, 24, 8, 0, 0, 0, time.UTC)},
	)

	out, err := service.ListAvailable(context.Background(), ListAvailableInput{Actor: shared.Actor{UserID: uuid.New(), Role: domain.RoleUser}, RoomID: roomID, Date: date})
	if err != nil {
		t.Fatalf("ListAvailable() error = %v", err)
	}

	if len(out.Slots) != len(wantSlots) {
		t.Fatalf("len(out.Slots) = %d, want %d", len(out.Slots), len(wantSlots))
	}

	if createManyCalled {
		t.Fatal("CreateMany should not be called when slots already exist")
	}
}

func TestService_ListAvailable_LazilyGeneratesSlots(t *testing.T) {
	t.Parallel()

	roomID := uuid.New()
	date := time.Date(2026, time.March, 24, 0, 0, 0, 0, time.UTC)
	schedule, err := domain.NewSchedule(uuid.New(), roomID, []domain.DayOfWeek{domain.Tuesday}, domain.TimeOfDay{Hour: 9}, domain.TimeOfDay{Hour: 10})
	if err != nil {
		t.Fatalf("NewSchedule() error = %v", err)
	}

	ids := &fakeIDGenerator{next: []uuid.UUID{uuid.New(), uuid.New()}}
	var createdSlots []domain.Slot
	checks := 0

	service := NewService(
		fakeRoomRepository{getByIDFn: func(_ context.Context, _ uuid.UUID) (domain.Room, bool, error) {
			return domain.Room{ID: roomID}, true, nil
		}},
		fakeScheduleRepository{getByRoomIDFn: func(_ context.Context, _ uuid.UUID) (domain.Schedule, bool, error) {
			return schedule, true, nil
		}},
		fakeSlotRepository{
			hasAnyByRoomAndDateFn: func(_ context.Context, _ uuid.UUID, _ time.Time) (bool, error) {
				checks++
				return false, nil
			},
			createManyFn: func(_ context.Context, slots []domain.Slot) error {
				createdSlots = append([]domain.Slot(nil), slots...)
				return nil
			},
			listAvailableFn: func(_ context.Context, _ uuid.UUID, _ time.Time) ([]domain.Slot, error) {
				return createdSlots, nil
			},
		},
		fakeTxManager{},
		ids,
		fakeClock{now: time.Date(2026, time.March, 24, 8, 0, 0, 0, time.UTC)},
	)

	out, err := service.ListAvailable(context.Background(), ListAvailableInput{Actor: shared.Actor{UserID: uuid.New(), Role: domain.RoleUser}, RoomID: roomID, Date: date})
	if err != nil {
		t.Fatalf("ListAvailable() error = %v", err)
	}

	if len(createdSlots) != 2 {
		t.Fatalf("len(createdSlots) = %d, want %d", len(createdSlots), 2)
	}

	if len(out.Slots) != 2 {
		t.Fatalf("len(out.Slots) = %d, want %d", len(out.Slots), 2)
	}

	if checks < 2 {
		t.Fatalf("HasAnyByRoomAndDate() checks = %d, want at least 2", checks)
	}
}

func TestService_ListAvailable_ReturnsValidationErrors(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.March, 24, 8, 0, 0, 0, time.UTC)
	service := NewService(fakeRoomRepository{}, fakeScheduleRepository{}, fakeSlotRepository{}, fakeTxManager{}, &fakeIDGenerator{}, fakeClock{now: now})

	tests := []struct {
		name  string
		input ListAvailableInput
		want  error
	}{
		{name: "unauthorized", input: ListAvailableInput{Actor: shared.Actor{}, RoomID: uuid.New(), Date: now}, want: shared.ErrUnauthorized},
		{name: "non utc date", input: ListAvailableInput{Actor: shared.Actor{UserID: uuid.New(), Role: domain.RoleUser}, RoomID: uuid.New(), Date: time.Date(2026, time.March, 24, 8, 0, 0, 0, time.FixedZone("UTC+3", 3*60*60))}, want: domain.ErrNonUTCTimestamp},
		{name: "too far ahead", input: ListAvailableInput{Actor: shared.Actor{UserID: uuid.New(), Role: domain.RoleUser}, RoomID: uuid.New(), Date: now.AddDate(0, 0, 31)}, want: shared.ErrDateTooFarAhead},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := service.ListAvailable(context.Background(), tt.input)
			if !errors.Is(err, tt.want) {
				t.Fatalf("ListAvailable() error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestService_ListAvailable_RoomNotFound(t *testing.T) {
	t.Parallel()

	service := NewService(
		fakeRoomRepository{getByIDFn: func(_ context.Context, _ uuid.UUID) (domain.Room, bool, error) {
			return domain.Room{}, false, nil
		}},
		fakeScheduleRepository{},
		fakeSlotRepository{},
		fakeTxManager{},
		&fakeIDGenerator{},
		fakeClock{now: time.Date(2026, time.March, 24, 8, 0, 0, 0, time.UTC)},
	)

	_, err := service.ListAvailable(context.Background(), ListAvailableInput{Actor: shared.Actor{UserID: uuid.New(), Role: domain.RoleAdmin}, RoomID: uuid.New(), Date: time.Date(2026, time.March, 24, 0, 0, 0, 0, time.UTC)})
	if !errors.Is(err, shared.ErrRoomNotFound) {
		t.Fatalf("ListAvailable() error = %v, want %v", err, shared.ErrRoomNotFound)
	}
}

var _ appports.RoomRepository = fakeRoomRepository{}
var _ appports.ScheduleRepository = fakeScheduleRepository{}
var _ appports.SlotRepository = fakeSlotRepository{}
var _ appports.TxManager = fakeTxManager{}
var _ appports.Clock = fakeClock{}
