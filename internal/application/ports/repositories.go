package ports

import (
	"context"
	"time"

	"github.com/avito-internships/test-backend-1-EdOoO21/internal/domain"
	"github.com/google/uuid"
)

// UserRepository хранит и читает пользователей
type UserRepository interface {
	// Create сохраняет нового пользователя вместе с хешем пароля.
	Create(ctx context.Context, user domain.User, passwordHash string) error
	// GetByID возвращает пользователя по идентификатору.
	GetByID(ctx context.Context, id uuid.UUID) (domain.User, bool, error)
	// GetByEmail возвращает пользователя и хеш его пароля по email.
	GetByEmail(ctx context.Context, email string) (domain.User, string, bool, error)
}

// RoomRepository хранит и читает переговорки
type RoomRepository interface {
	// Create сохраняет новую переговорку
	Create(ctx context.Context, room domain.Room) error
	// GetByID возвращает переговорку по идентификатору
	GetByID(ctx context.Context, id uuid.UUID) (domain.Room, bool, error)
	// List возвращает список всех переговорок
	List(ctx context.Context) ([]domain.Room, error)
}

// ScheduleRepository хранит расписания переговорок
type ScheduleRepository interface {
	// Create сохраняет расписание переговорки
	Create(ctx context.Context, schedule domain.Schedule) error
	// GetByRoomID возвращает расписание переговорки
	GetByRoomID(ctx context.Context, roomID uuid.UUID) (domain.Schedule, bool, error)
}

// SlotRepository хранит и читает слоты бронирования
type SlotRepository interface {
	// CreateMany сохраняет набор сгенерированных слотов
	CreateMany(ctx context.Context, slots []domain.Slot) error
	// GetByID возвращает слот по идентификатору
	GetByID(ctx context.Context, id uuid.UUID) (domain.Slot, bool, error)
	// HasAnyByRoomAndDate проверяет, есть ли у переговорки хоть один слот на дату
	HasAnyByRoomAndDate(ctx context.Context, roomID uuid.UUID, date time.Time) (bool, error)
	// ListAvailableByRoomAndDate возвращает свободные слоты переговорки на дату
	ListAvailableByRoomAndDate(ctx context.Context, roomID uuid.UUID, date time.Time) ([]domain.Slot, error)
}

// BookingRepository хранит и читает бронирования
type BookingRepository interface {
	// Create сохраняет новую бронь
	Create(ctx context.Context, booking domain.Booking) error
	// Update обновляет существующую бронь
	Update(ctx context.Context, booking domain.Booking) error
	// GetByID возвращает бронь по идентификатору
	GetByID(ctx context.Context, id uuid.UUID) (domain.Booking, bool, error)
	// HasActiveBySlotID проверяет, есть ли у слота активная бронь
	HasActiveBySlotID(ctx context.Context, slotID uuid.UUID) (bool, error)
	// ListByUserFuture возвращает будущие брони пользователя
	ListByUserFuture(ctx context.Context, userID uuid.UUID, now time.Time) ([]domain.Booking, error)
	// List возвращает страницу всех броней и их общее количество
	List(ctx context.Context, page, pageSize int) ([]domain.Booking, int, error)
}
