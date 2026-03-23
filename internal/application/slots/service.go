package slots

import (
	"context"
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
		return ListAvailableOutput{}, err
	}

	date, err := domain.RequireUTC(input.Date)
	if err != nil {
		return ListAvailableOutput{}, err
	}

	date = startOfDayUTC(date)
	now := startOfDayUTC(s.clock.NowUTC())

	if date.After(now.AddDate(0, 0, schedules.MaxSlotGenerationDaysAhead)) {
		return ListAvailableOutput{}, shared.ErrDateTooFarAhead
	}

	_, exists, err := s.rooms.GetByID(ctx, input.RoomID)
	if err != nil {
		return ListAvailableOutput{}, err
	}
	if !exists {
		return ListAvailableOutput{}, shared.ErrRoomNotFound
	}

	if err := s.ensureSlotsForDate(ctx, input.RoomID, date); err != nil {
		return ListAvailableOutput{}, err
	}

	slots, err := s.slots.ListAvailableByRoomAndDate(ctx, input.RoomID, date)
	if err != nil {
		return ListAvailableOutput{}, err
	}

	return ListAvailableOutput{Slots: slots}, nil
}

func (s *Service) ensureSlotsForDate(ctx context.Context, roomID uuid.UUID, date time.Time) error {
	hasAny, err := s.slots.HasAnyByRoomAndDate(ctx, roomID, date)
	if err != nil {
		return err
	}
	if hasAny {
		return nil
	}

	schedule, exists, err := s.schedules.GetByRoomID(ctx, roomID)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	generated, err := schedules.GenerateSlotsForDate(s.ids, schedule, date)
	if err != nil {
		return err
	}
	if len(generated) == 0 {
		return nil
	}

	return s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		hasAny, err := s.slots.HasAnyByRoomAndDate(txCtx, roomID, date)
		if err != nil {
			return err
		}
		if hasAny {
			return nil
		}

		return s.slots.CreateMany(txCtx, generated)
	})
}

func startOfDayUTC(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}
