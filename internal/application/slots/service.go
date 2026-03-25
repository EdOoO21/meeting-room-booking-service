package slots

import (
	"context"
	"fmt"
	"time"

	appports "github.com/avito-internships/test-backend-1-EdOoO21/internal/application/ports"
	"github.com/avito-internships/test-backend-1-EdOoO21/internal/application/schedules"
	"github.com/avito-internships/test-backend-1-EdOoO21/internal/application/shared"
	"github.com/avito-internships/test-backend-1-EdOoO21/internal/domain"
	"github.com/google/uuid"
)

type Service struct {
	rooms     appports.RoomRepository
	schedules appports.ScheduleRepository
	slots     appports.SlotRepository
	tx        appports.TxManager
	ids       appports.IDGenerator
	clock     appports.Clock
}

type ListAvailableInput struct {
	Actor  shared.Actor
	RoomID uuid.UUID
	Date   time.Time
}

type ListAvailableOutput struct {
	Slots []domain.Slot
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

func (s *Service) ListAvailable(ctx context.Context, input ListAvailableInput) (ListAvailableOutput, error) {
	if err := input.Actor.RequireRole(domain.RoleAdmin, domain.RoleUser); err != nil {
		return ListAvailableOutput{}, fmt.Errorf("authorize available slots list: %w", err)
	}

	date, err := domain.RequireUTC(input.Date)
	if err != nil {
		return ListAvailableOutput{}, fmt.Errorf("normalize requested date: %w", err)
	}

	date = startOfDayUTC(date)
	now := startOfDayUTC(s.clock.NowUTC())

	if date.After(now.AddDate(0, 0, schedules.MaxSlotGenerationDaysAhead)) {
		return ListAvailableOutput{}, shared.ErrDateTooFarAhead
	}

	_, exists, err := s.rooms.GetByID(ctx, input.RoomID)
	if err != nil {
		return ListAvailableOutput{}, fmt.Errorf("get room by id: %w", err)
	}
	if !exists {
		return ListAvailableOutput{}, shared.ErrRoomNotFound
	}

	if ensureErr := s.ensureSlotsForDate(ctx, input.RoomID, date); ensureErr != nil {
		return ListAvailableOutput{}, ensureErr
	}

	slots, err := s.slots.ListAvailableByRoomAndDate(ctx, input.RoomID, date)
	if err != nil {
		return ListAvailableOutput{}, fmt.Errorf("list available slots by room and date: %w", err)
	}

	return ListAvailableOutput{Slots: slots}, nil
}

func (s *Service) ensureSlotsForDate(ctx context.Context, roomID uuid.UUID, date time.Time) error {
	hasAny, err := s.slots.HasAnyByRoomAndDate(ctx, roomID, date)
	if err != nil {
		return fmt.Errorf("check slots for date: %w", err)
	}
	if hasAny {
		return nil
	}

	schedule, exists, err := s.schedules.GetByRoomID(ctx, roomID)
	if err != nil {
		return fmt.Errorf("get schedule by room id: %w", err)
	}
	if !exists {
		return nil
	}

	generated, err := schedules.GenerateSlotsForDate(s.ids, schedule, date)
	if err != nil {
		return fmt.Errorf("generate slots for date: %w", err)
	}
	if len(generated) == 0 {
		return nil
	}

	if txErr := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		hasAnyInTx, hasAnyErr := s.slots.HasAnyByRoomAndDate(txCtx, roomID, date)
		if hasAnyErr != nil {
			return fmt.Errorf("check slots for date in transaction: %w", hasAnyErr)
		}
		if hasAnyInTx {
			return nil
		}

		if createManyErr := s.slots.CreateMany(txCtx, generated); createManyErr != nil {
			return fmt.Errorf("create generated slots: %w", createManyErr)
		}

		return nil
	}); txErr != nil {
		return fmt.Errorf("ensure slots for date transaction: %w", txErr)
	}

	return nil
}

func startOfDayUTC(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}
