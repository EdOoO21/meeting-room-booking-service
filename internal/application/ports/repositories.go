package ports

import (
	"context"
	"time"

	"github.com/avito-internships/test-backend-1-EdOoO21/internal/domain"
	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, user domain.User) error
	GetByID(ctx context.Context, id uuid.UUID) (domain.User, bool, error)
	GetByEmail(ctx context.Context, email string) (domain.User, bool, error)
}

type RoomRepository interface {
	Create(ctx context.Context, room domain.Room) error
	GetByID(ctx context.Context, id uuid.UUID) (domain.Room, bool, error)
	List(ctx context.Context) ([]domain.Room, error)
}

type ScheduleRepository interface {
	Create(ctx context.Context, schedule domain.Schedule) error
	GetByRoomID(ctx context.Context, roomID uuid.UUID) (domain.Schedule, bool, error)
}

type SlotRepository interface {
	CreateMany(ctx context.Context, slots []domain.Slot) error
	GetByID(ctx context.Context, id uuid.UUID) (domain.Slot, bool, error)
	ListAvailableByRoomAndDate(ctx context.Context, roomID uuid.UUID, date time.Time) ([]domain.Slot, error)
}

type BookingRepository interface {
	Create(ctx context.Context, booking domain.Booking) error
	Update(ctx context.Context, booking domain.Booking) error
	GetByID(ctx context.Context, id uuid.UUID) (domain.Booking, bool, error)
	HasActiveBySlotID(ctx context.Context, slotID uuid.UUID) (bool, error)
	ListByUserFuture(ctx context.Context, userID uuid.UUID, now time.Time) ([]domain.Booking, error)
	List(ctx context.Context, page, pageSize int) ([]domain.Booking, int, error)
}
