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

type ScheduleRepository struct {
	db *DB
}

func NewScheduleRepository(db *DB) *ScheduleRepository {
	return &ScheduleRepository{db: db}
}

func (r *ScheduleRepository) Create(ctx context.Context, schedule domain.Schedule) error {
	const query = "INSERT INTO schedules (id, room_id, days_of_week, start_time, end_time) VALUES (@id, @room_id, @days_of_week, @start_time, @end_time)"

	_, err := r.db.querier(ctx).Exec(ctx, query, pgx.NamedArgs{
		"id":           schedule.ID,
		"room_id":      schedule.RoomID,
		"days_of_week": dbDaysOfWeek(schedule.DaysOfWeek),
		"start_time":   timeOfDayDBValue(schedule.StartTime),
		"end_time":     timeOfDayDBValue(schedule.EndTime),
	})
	if err != nil {
		return fmt.Errorf("create schedule: %w", err)
	}

	return nil
}

func (r *ScheduleRepository) GetByRoomID(ctx context.Context, roomID uuid.UUID) (domain.Schedule, bool, error) {
	const query = "SELECT id, room_id, days_of_week, start_time, end_time FROM schedules WHERE room_id = @room_id"
	row := r.db.querier(ctx).QueryRow(ctx, query, pgx.NamedArgs{"room_id": roomID})

	var (
		id        uuid.UUID
		loadedID  uuid.UUID
		days      []int16
		startTime time.Time
		endTime   time.Time
	)

	if err := row.Scan(&id, &loadedID, &days, &startTime, &endTime); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Schedule{}, false, nil
		}
		return domain.Schedule{}, false, wrapScanError("schedule", err)
	}

	domainDays := scanDaysOfWeek(days)

	domainStart, err := scanTimeOfDay(startTime)
	if err != nil {
		return domain.Schedule{}, false, fmt.Errorf("build schedule start time: %w", err)
	}

	domainEnd, err := scanTimeOfDay(endTime)
	if err != nil {
		return domain.Schedule{}, false, fmt.Errorf("build schedule end time: %w", err)
	}

	schedule, err := domain.NewSchedule(id, loadedID, domainDays, domainStart, domainEnd)
	if err != nil {
		return domain.Schedule{}, false, fmt.Errorf("build schedule: %w", err)
	}

	return schedule, true, nil
}
