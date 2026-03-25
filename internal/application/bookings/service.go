package bookings

import (
	"context"
	"fmt"

	appports "github.com/avito-internships/test-backend-1-EdOoO21/internal/application/ports"
	"github.com/avito-internships/test-backend-1-EdOoO21/internal/application/shared"
	"github.com/avito-internships/test-backend-1-EdOoO21/internal/domain"
	"github.com/google/uuid"
)

type Service struct {
	bookings        appports.BookingRepository
	slots           appports.SlotRepository
	tx              appports.TxManager
	ids             appports.IDGenerator
	clock           appports.Clock
	conferenceLinks appports.ConferenceLinkService
}

type CreateInput struct {
	Actor                shared.Actor
	SlotID               uuid.UUID
	CreateConferenceLink bool
}

type CreateOutput struct {
	Booking domain.Booking
}

type CancelInput struct {
	Actor     shared.Actor
	BookingID uuid.UUID
}

type CancelOutput struct {
	Booking domain.Booking
}

type ListMineInput struct {
	Actor shared.Actor
}

type ListMineOutput struct {
	Bookings []domain.Booking
}

type ListInput struct {
	Actor    shared.Actor
	Page     int
	PageSize int
}

type ListOutput struct {
	Bookings   []domain.Booking
	Pagination shared.Pagination
}

func NewService(
	bookings appports.BookingRepository,
	slots appports.SlotRepository,
	tx appports.TxManager,
	ids appports.IDGenerator,
	clock appports.Clock,
	conferenceLinks appports.ConferenceLinkService,
) *Service {
	return &Service{
		bookings:        bookings,
		slots:           slots,
		tx:              tx,
		ids:             ids,
		clock:           clock,
		conferenceLinks: conferenceLinks,
	}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (CreateOutput, error) {
	if err := input.Actor.RequireRole(domain.RoleUser); err != nil {
		return CreateOutput{}, fmt.Errorf("authorize booking creation: %w", err)
	}

	now := s.clock.NowUTC()
	bookingID := s.ids.NewUUID()
	var created domain.Booking

	if err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		slot, exists, err := s.slots.GetByID(txCtx, input.SlotID)
		if err != nil {
			return fmt.Errorf("get slot by id: %w", err)
		}
		if !exists {
			return shared.ErrSlotNotFound
		}

		busy, err := s.bookings.HasActiveBySlotID(txCtx, input.SlotID)
		if err != nil {
			return fmt.Errorf("check slot booking status: %w", err)
		}
		if busy {
			return shared.ErrSlotBooked
		}

		var conferenceLink *string
		if input.CreateConferenceLink && s.conferenceLinks != nil {
			link, linkErr := s.conferenceLinks.CreateConferenceLink(txCtx, bookingID)
			if linkErr != nil {
				return fmt.Errorf("create conference link: %w", linkErr)
			}
			conferenceLink = &link
		}

		booking, err := domain.NewActiveBooking(bookingID, input.SlotID, input.Actor.UserID, conferenceLink, now)
		if err != nil {
			return fmt.Errorf("build active booking: %w", err)
		}

		if validationErr := booking.CanBeCreatedBy(input.Actor.Role, slot, now); validationErr != nil {
			return fmt.Errorf("validate booking creation: %w", validationErr)
		}

		if createErr := s.bookings.Create(txCtx, booking); createErr != nil {
			return fmt.Errorf("create booking: %w", createErr)
		}

		created = booking
		return nil
	}); err != nil {
		return CreateOutput{}, fmt.Errorf("create booking transaction: %w", err)
	}

	return CreateOutput{Booking: created}, nil
}

func (s *Service) Cancel(ctx context.Context, input CancelInput) (CancelOutput, error) {
	if err := input.Actor.RequireRole(domain.RoleUser); err != nil {
		return CancelOutput{}, fmt.Errorf("authorize booking cancellation: %w", err)
	}

	var result domain.Booking

	if err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		booking, exists, err := s.bookings.GetByID(txCtx, input.BookingID)
		if err != nil {
			return fmt.Errorf("get booking by id: %w", err)
		}
		if !exists {
			return shared.ErrBookingNotFound
		}

		if !booking.BelongsTo(input.Actor.UserID) {
			return shared.ErrForbidden
		}

		if booking.IsActive() {
			booking.Cancel()
			if updateErr := s.bookings.Update(txCtx, booking); updateErr != nil {
				return fmt.Errorf("update booking: %w", updateErr)
			}
		}

		result = booking
		return nil
	}); err != nil {
		return CancelOutput{}, fmt.Errorf("cancel booking transaction: %w", err)
	}

	return CancelOutput{Booking: result}, nil
}

func (s *Service) ListMine(ctx context.Context, input ListMineInput) (ListMineOutput, error) {
	if err := input.Actor.RequireRole(domain.RoleUser); err != nil {
		return ListMineOutput{}, fmt.Errorf("authorize own bookings list: %w", err)
	}

	bookings, err := s.bookings.ListByUserFuture(ctx, input.Actor.UserID, s.clock.NowUTC())
	if err != nil {
		return ListMineOutput{}, fmt.Errorf("list future bookings for user: %w", err)
	}

	return ListMineOutput{Bookings: bookings}, nil
}

func (s *Service) List(ctx context.Context, input ListInput) (ListOutput, error) {
	if err := input.Actor.RequireRole(domain.RoleAdmin); err != nil {
		return ListOutput{}, fmt.Errorf("authorize bookings list: %w", err)
	}

	if input.Page < 1 {
		return ListOutput{}, shared.ErrInvalidPage
	}

	if input.PageSize < 1 {
		return ListOutput{}, shared.ErrInvalidPageSize
	}

	bookings, total, err := s.bookings.List(ctx, input.Page, input.PageSize)
	if err != nil {
		return ListOutput{}, fmt.Errorf("list bookings: %w", err)
	}

	return ListOutput{
		Bookings: bookings,
		Pagination: shared.Pagination{
			Page:     input.Page,
			PageSize: input.PageSize,
			Total:    total,
		},
	}, nil
}
