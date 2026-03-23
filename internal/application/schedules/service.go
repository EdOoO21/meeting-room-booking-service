package schedules

import (
	"context"
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
		return CreateOutput{}, err
	}

	startTime, err := domain.ParseTimeOfDay(input.StartTime)
	if err != nil {
		return CreateOutput{}, err
	}

	endTime, err := domain.ParseTimeOfDay(input.EndTime)
	if err != nil {
		return CreateOutput{}, err
	}

	schedule, err := domain.NewSchedule(s.ids.NewUUID(), input.RoomID, input.DaysOfWeek, startTime, endTime)
	if err != nil {
		return CreateOutput{}, err
	}

	now := s.clock.NowUTC()
	generatedSlots, err := s.generateUpcomingSlots(schedule, now, InitialSlotGenerationDays)
	if err != nil {
		return CreateOutput{}, err
	}

	if err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		_, exists, err := s.rooms.GetByID(txCtx, input.RoomID)
		if err != nil {
			return err
		}
		if !exists {
			return shared.ErrRoomNotFound
		}

		_, exists, err = s.schedules.GetByRoomID(txCtx, input.RoomID)
		if err != nil {
			return err
		}
		if exists {
			return shared.ErrScheduleExists
		}

		if err := s.schedules.Create(txCtx, schedule); err != nil {
			return err
		}

		if len(generatedSlots) == 0 {
			return nil
		}

		return s.slots.CreateMany(txCtx, generatedSlots)
	}); err != nil {
		return CreateOutput{}, err
	}

	return CreateOutput{Schedule: schedule, GeneratedSlots: len(generatedSlots)}, nil
}

func GenerateSlotsForDate(ids appports.IDGenerator, schedule domain.Schedule, date time.Time) ([]domain.Slot, error) {
	targetDate, err := domain.RequireUTC(date)
	if err != nil {
		return nil, err
	}

	dayStart := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, time.UTC)
	if !scheduleIncludesDate(schedule, dayStart) {
		return []domain.Slot{}, nil
	}

	slots := make([]domain.Slot, 0)
	currentStart := dayStart.Add(schedule.StartTime.ToDuration())
	windowEnd := dayStart.Add(schedule.EndTime.ToDuration())

	for !currentStart.Add(domain.SlotDuration).After(windowEnd) {
		slot, err := domain.NewSlot(ids.NewUUID(), schedule.RoomID, currentStart, currentStart.Add(domain.SlotDuration))
		if err != nil {
			return nil, err
		}

		slots = append(slots, slot)
		currentStart = currentStart.Add(domain.SlotDuration)
	}

	return slots, nil
}

func (s *Service) generateUpcomingSlots(schedule domain.Schedule, from time.Time, days int) ([]domain.Slot, error) {
	baseTime, err := domain.RequireUTC(from.UTC())
	if err != nil {
		return nil, err
	}

	dayStart := time.Date(baseTime.Year(), baseTime.Month(), baseTime.Day(), 0, 0, 0, 0, time.UTC)
	slots := make([]domain.Slot, 0)

	for offset := range days {
		date := dayStart.AddDate(0, 0, offset)
		generated, err := GenerateSlotsForDate(s.ids, schedule, date)
		if err != nil {
			return nil, err
		}

		slots = append(slots, generated...)
	}

	return slots, nil
}

func scheduleIncludesDate(schedule domain.Schedule, date time.Time) bool {
	weekday := domain.DayOfWeekFromWeekday(date.Weekday())
	return slices.Contains(schedule.DaysOfWeek, weekday)
}
