package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/avito-internships/test-backend-1-EdOoO21/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type SlotRepository struct {
	db *DB
}

func NewSlotRepository(db *DB) *SlotRepository {
	return &SlotRepository{db: db}
}

func (r *SlotRepository) CreateMany(ctx context.Context, slots []domain.Slot) error {
	const query = "INSERT INTO slots (id, room_id, start_at, end_at) VALUES (@id, @room_id, @start_at, @end_at) ON CONFLICT (room_id, start_at) DO NOTHING"

	if len(slots) == 0 {
		return nil
	}

	for _, slot := range slots {
		if _, err := r.db.querier(ctx).Exec(ctx, query, pgx.NamedArgs{
			"id":       slot.ID,
			"room_id":  slot.RoomID,
			"start_at": slot.Start,
			"end_at":   slot.End,
		}); err != nil {
			return fmt.Errorf("create slot: %w", err)
		}
	}

	return nil
}

func (r *SlotRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.Slot, bool, error) {
	const query = "SELECT id, room_id, start_at, end_at FROM slots WHERE id = @id"
	row := r.db.querier(ctx).QueryRow(ctx, query, pgx.NamedArgs{"id": id})

	return scanSlot(row)
}

func (r *SlotRepository) HasAnyByRoomAndDate(ctx context.Context, roomID uuid.UUID, date time.Time) (bool, error) {
	const query = "SELECT EXISTS (SELECT 1 FROM slots WHERE room_id = @room_id AND start_at >= @start_at AND start_at < @end_at)"

	dayStart := startOfDayUTC(date)
	dayEnd := dayStart.AddDate(0, 0, 1)

	var exists bool
	if err := r.db.querier(ctx).QueryRow(ctx, query, pgx.NamedArgs{
		"room_id":  roomID,
		"start_at": dayStart,
		"end_at":   dayEnd,
	}).Scan(&exists); err != nil {
		return false, fmt.Errorf("check slots by room and date: %w", err)
	}

	return exists, nil
}

func (r *SlotRepository) ListAvailableByRoomAndDate(ctx context.Context, roomID uuid.UUID, date time.Time) ([]domain.Slot, error) {
	const query = "SELECT s.id, s.room_id, s.start_at, s.end_at FROM slots s LEFT JOIN bookings b ON b.slot_id = s.id AND b.status = 'active' WHERE s.room_id = @room_id AND s.start_at >= @start_at AND s.start_at < @end_at AND b.id IS NULL ORDER BY s.start_at ASC"

	dayStart := startOfDayUTC(date)
	dayEnd := dayStart.AddDate(0, 0, 1)

	rows, err := r.db.querier(ctx).Query(ctx, query, pgx.NamedArgs{
		"room_id":  roomID,
		"start_at": dayStart,
		"end_at":   dayEnd,
	})
	if err != nil {
		return nil, fmt.Errorf("list available slots: %w", err)
	}
	defer rows.Close()

	slots := make([]domain.Slot, 0)
	for rows.Next() {
		slot, scanErr := scanSlotRow(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		slots = append(slots, slot)
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, fmt.Errorf("iterate slots: %w", rowsErr)
	}

	return slots, nil
}

func scanSlot(row pgx.Row) (domain.Slot, bool, error) {
	slot, err := scanSlotRow(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Slot{}, false, nil
		}
		return domain.Slot{}, false, err
	}

	return slot, true, nil
}

func scanSlotRow(row interface{ Scan(dest ...any) error }) (domain.Slot, error) {
	var (
		id      uuid.UUID
		roomID  uuid.UUID
		startAt time.Time
		endAt   time.Time
	)

	if err := row.Scan(&id, &roomID, &startAt, &endAt); err != nil {
		return domain.Slot{}, wrapScanError("slot", err)
	}

	slot, err := domain.NewSlot(id, roomID, normalizeScannedTimestamp(startAt), normalizeScannedTimestamp(endAt))
	if err != nil {
		return domain.Slot{}, fmt.Errorf("build slot: %w", err)
	}

	return slot, nil
}
