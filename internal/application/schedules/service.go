package schedules

import (
	"context"
	"fmt"
	"slices"
	"time"

	appports "github.com/avito-internships/test-backend-1-EdOoO21/internal/application/ports"
	"github.com/avito-internships/test-backend-1-EdOoO21/internal/application/shared"
	"github.com/avito-internships/test-backend-1-EdOoO21/internal/domain"
	"github.com/google/uuid"
)

const InitialSlotGenerationDays = 7
const MaxSlotGenerationDaysAhead = 30

type Service struct {
	rooms     appports.RoomRepository
	schedules appports.ScheduleRepository
	slots     appports.SlotRepository
	tx        appports.TxManager
	ids       appports.IDGenerator
	clock     appports.Clock
}

type CreateInput struct {
	Actor      shared.Actor
	RoomID     uuid.UUID
	DaysOfWeek []domain.DayOfWeek
	StartTime  string
	EndTime    string
}

type CreateOutput struct {
	Schedule       domain.Schedule
	GeneratedSlots int
}

func NewService(
	rooms appports.RoomRepository,
	schedules appports.ScheduleRepository,
	slots appports.SlotRepository,
	tx appports.TxManager,
	ids appports.IDGenerator,
	clock appports.Clock,
) *Service {
	return &Service{rooms: rooms, schedules: schedules, slots: slots, tx: tx, ids: ids, clock: clock}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (CreateOutput, error) {
	if err := input.Actor.RequireRole(domain.RoleAdmin); err != nil {
		return CreateOutput{}, fmt.Errorf("authorize schedule creation: %w", err)
	}

	startTime, err := domain.ParseTimeOfDay(input.StartTime)
	if err != nil {
		return CreateOutput{}, fmt.Errorf("parse schedule start time: %w", err)
	}

	endTime, err := domain.ParseTimeOfDay(input.EndTime)
	if err != nil {
		return CreateOutput{}, fmt.Errorf("parse schedule end time: %w", err)
	}

	schedule, err := domain.NewSchedule(s.ids.NewUUID(), input.RoomID, input.DaysOfWeek, startTime, endTime)
	if err != nil {
		return CreateOutput{}, fmt.Errorf("build schedule: %w", err)
	}

	now := s.clock.NowUTC()
	generatedSlots, err := s.generateUpcomingSlots(schedule, now, InitialSlotGenerationDays)
	if err != nil {
		return CreateOutput{}, fmt.Errorf("generate initial slots: %w", err)
	}

	if txErr := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		_, exists, roomErr := s.rooms.GetByID(txCtx, input.RoomID)
		if roomErr != nil {
			return fmt.Errorf("get room by id: %w", roomErr)
		}
		if !exists {
			return shared.ErrRoomNotFound
		}

		_, exists, err = s.schedules.GetByRoomID(txCtx, input.RoomID)
		if err != nil {
			return fmt.Errorf("get schedule by room id: %w", err)
		}
		if exists {
			return shared.ErrScheduleExists
		}

		if createErr := s.schedules.Create(txCtx, schedule); createErr != nil {
			return fmt.Errorf("create schedule: %w", createErr)
		}

		if len(generatedSlots) == 0 {
			return nil
		}

		if createManyErr := s.slots.CreateMany(txCtx, generatedSlots); createManyErr != nil {
			return fmt.Errorf("create generated slots: %w", createManyErr)
		}

		return nil
	}); txErr != nil {
		return CreateOutput{}, fmt.Errorf("create schedule transaction: %w", txErr)
	}

	return CreateOutput{Schedule: schedule, GeneratedSlots: len(generatedSlots)}, nil
}

func GenerateSlotsForDate(ids appports.IDGenerator, schedule domain.Schedule, date time.Time) ([]domain.Slot, error) {
	targetDate, err := domain.RequireUTC(date)
	if err != nil {
		return nil, fmt.Errorf("require UTC date: %w", err)
	}

	dayStart := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, time.UTC)
	if !scheduleIncludesDate(schedule, dayStart) {
		return []domain.Slot{}, nil
	}

	slots := make([]domain.Slot, 0)
	currentStart := dayStart.Add(schedule.StartTime.ToDuration())
	windowEnd := dayStart.Add(schedule.EndTime.ToDuration())

	for !currentStart.Add(domain.SlotDuration).After(windowEnd) {
		slot, slotErr := domain.NewSlot(ids.NewUUID(), schedule.RoomID, currentStart, currentStart.Add(domain.SlotDuration))
		if slotErr != nil {
			return nil, fmt.Errorf("build slot for date %s: %w", dayStart.Format(time.DateOnly), slotErr)
		}

		slots = append(slots, slot)
		currentStart = currentStart.Add(domain.SlotDuration)
	}

	return slots, nil
}

func (s *Service) generateUpcomingSlots(schedule domain.Schedule, from time.Time, days int) ([]domain.Slot, error) {
	baseTime, err := domain.RequireUTC(from.UTC())
	if err != nil {
		return nil, fmt.Errorf("require UTC base time: %w", err)
	}

	dayStart := time.Date(baseTime.Year(), baseTime.Month(), baseTime.Day(), 0, 0, 0, 0, time.UTC)
	slots := make([]domain.Slot, 0)

	for offset := range days {
		date := dayStart.AddDate(0, 0, offset)
		generated, generatedErr := GenerateSlotsForDate(s.ids, schedule, date)
		if generatedErr != nil {
			return nil, fmt.Errorf("generate slots for %s: %w", date.Format(time.DateOnly), generatedErr)
		}

		slots = append(slots, generated...)
	}

	return slots, nil
}

func scheduleIncludesDate(schedule domain.Schedule, date time.Time) bool {
	weekday := domain.DayOfWeekFromWeekday(date.Weekday())
	return slices.Contains(schedule.DaysOfWeek, weekday)
}
