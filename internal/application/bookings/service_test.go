package bookings

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

type fakeBookingRepository struct {
	createFn            func(_ context.Context, booking domain.Booking) error
	updateFn            func(_ context.Context, booking domain.Booking) error
	getByIDFn           func(_ context.Context, id uuid.UUID) (domain.Booking, bool, error)
	hasActiveBySlotIDFn func(_ context.Context, slotID uuid.UUID) (bool, error)
	listByUserFutureFn  func(_ context.Context, userID uuid.UUID, now time.Time) ([]domain.Booking, error)
	listFn              func(_ context.Context, page, pageSize int) ([]domain.Booking, int, error)
}

func (f fakeBookingRepository) Create(ctx context.Context, booking domain.Booking) error {
	if f.createFn != nil {
		return f.createFn(ctx, booking)
	}
	return nil
}

func (f fakeBookingRepository) Update(ctx context.Context, booking domain.Booking) error {
	if f.updateFn != nil {
		return f.updateFn(ctx, booking)
	}
	return nil
}

func (f fakeBookingRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.Booking, bool, error) {
	if f.getByIDFn != nil {
		return f.getByIDFn(ctx, id)
	}
	return domain.Booking{}, false, nil
}

func (f fakeBookingRepository) HasActiveBySlotID(ctx context.Context, slotID uuid.UUID) (bool, error) {
	if f.hasActiveBySlotIDFn != nil {
		return f.hasActiveBySlotIDFn(ctx, slotID)
	}
	return false, nil
}

func (f fakeBookingRepository) ListByUserFuture(ctx context.Context, userID uuid.UUID, now time.Time) ([]domain.Booking, error) {
	if f.listByUserFutureFn != nil {
		return f.listByUserFutureFn(ctx, userID, now)
	}
	return nil, nil
}

func (f fakeBookingRepository) List(ctx context.Context, page, pageSize int) ([]domain.Booking, int, error) {
	if f.listFn != nil {
		return f.listFn(ctx, page, pageSize)
	}
	return nil, 0, nil
}

type fakeSlotRepository struct {
	createManyFn          func(_ context.Context, slots []domain.Slot) error
	getByIDFn             func(_ context.Context, id uuid.UUID) (domain.Slot, bool, error)
	hasAnyByRoomAndDateFn func(_ context.Context, roomID uuid.UUID, date time.Time) (bool, error)
	listAvailableFn       func(_ context.Context, roomID uuid.UUID, date time.Time) ([]domain.Slot, error)
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
	withinTransactionFn func(_ context.Context, fn func(ctx context.Context) error) error
}

func (f fakeTxManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	if f.withinTransactionFn != nil {
		return f.withinTransactionFn(ctx, fn)
	}
	return fn(ctx)
}

type fakeIDGenerator struct{ next uuid.UUID }

func (f fakeIDGenerator) NewUUID() uuid.UUID { return f.next }

type fakeClock struct{ now time.Time }

func (f fakeClock) NowUTC() time.Time { return f.now }

type fakeConferenceLinkService struct {
	createFn func(_ context.Context, bookingID uuid.UUID) (string, error)
}

func (f fakeConferenceLinkService) CreateConferenceLink(ctx context.Context, bookingID uuid.UUID) (string, error) {
	if f.createFn != nil {
		return f.createFn(ctx, bookingID)
	}
	return "", nil
}

func TestService_Create_Success(t *testing.T) {
	t.Parallel()

	bookingID := uuid.New()
	slotID := uuid.New()
	userID := uuid.New()
	now := time.Date(2026, time.March, 24, 12, 0, 0, 0, time.UTC)
	slot, err := domain.NewSlot(uuid.New(), uuid.New(), now.Add(time.Hour), now.Add(time.Hour).Add(domain.SlotDuration))
	if err != nil {
		t.Fatalf("NewSlot() error = %v", err)
	}

	var created domain.Booking
	service := NewService(
		fakeBookingRepository{
			hasActiveBySlotIDFn: func(_ context.Context, _ uuid.UUID) (bool, error) { return false, nil },
			createFn: func(_ context.Context, booking domain.Booking) error {
				created = booking
				return nil
			},
		},
		fakeSlotRepository{getByIDFn: func(_ context.Context, _ uuid.UUID) (domain.Slot, bool, error) {
			return slot, true, nil
		}},
		fakeTxManager{},
		fakeIDGenerator{next: bookingID},
		fakeClock{now: now},
		fakeConferenceLinkService{createFn: func(_ context.Context, gotBookingID uuid.UUID) (string, error) {
			if gotBookingID != bookingID {
				t.Fatalf("conference bookingID = %v, want %v", gotBookingID, bookingID)
			}
			return "https://meet.example/abc", nil
		}},
	)

	out, err := service.Create(context.Background(), CreateInput{Actor: shared.Actor{UserID: userID, Role: domain.RoleUser}, SlotID: slotID, CreateConferenceLink: true})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if out.Booking.ID != bookingID {
		t.Fatalf("out.Booking.ID = %v, want %v", out.Booking.ID, bookingID)
	}

	if created.ConferenceLink == nil || *created.ConferenceLink != "https://meet.example/abc" {
		t.Fatalf("created.ConferenceLink = %v, want conference link", created.ConferenceLink)
	}
}

func TestService_Create_ReturnsExpectedErrors(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.March, 24, 12, 0, 0, 0, time.UTC)
	pastSlot, err := domain.NewSlot(uuid.New(), uuid.New(), now.Add(-time.Hour), now.Add(-time.Hour).Add(domain.SlotDuration))
	if err != nil {
		t.Fatalf("NewSlot() error = %v", err)
	}

	tests := []struct {
		name    string
		service *Service
		input   CreateInput
		want    error
	}{
		{
			name:    "forbidden role",
			service: NewService(fakeBookingRepository{}, fakeSlotRepository{}, fakeTxManager{}, fakeIDGenerator{next: uuid.New()}, fakeClock{now: now}, nil),
			input:   CreateInput{Actor: shared.Actor{UserID: uuid.New(), Role: domain.RoleAdmin}, SlotID: uuid.New()},
			want:    shared.ErrForbidden,
		},
		{
			name: "slot not found",
			service: NewService(
				fakeBookingRepository{},
				fakeSlotRepository{getByIDFn: func(_ context.Context, _ uuid.UUID) (domain.Slot, bool, error) { return domain.Slot{}, false, nil }},
				fakeTxManager{},
				fakeIDGenerator{next: uuid.New()},
				fakeClock{now: now},
				nil,
			),
			input: CreateInput{Actor: shared.Actor{UserID: uuid.New(), Role: domain.RoleUser}, SlotID: uuid.New()},
			want:  shared.ErrSlotNotFound,
		},
		{
			name: "slot booked",
			service: NewService(
				fakeBookingRepository{hasActiveBySlotIDFn: func(_ context.Context, _ uuid.UUID) (bool, error) { return true, nil }},
				fakeSlotRepository{getByIDFn: func(_ context.Context, _ uuid.UUID) (domain.Slot, bool, error) {
					future := time.Date(2026, time.March, 24, 13, 0, 0, 0, time.UTC)
					slot, _ := domain.NewSlot(uuid.New(), uuid.New(), future, future.Add(domain.SlotDuration))
					return slot, true, nil
				}},
				fakeTxManager{},
				fakeIDGenerator{next: uuid.New()},
				fakeClock{now: now},
				nil,
			),
			input: CreateInput{Actor: shared.Actor{UserID: uuid.New(), Role: domain.RoleUser}, SlotID: uuid.New()},
			want:  shared.ErrSlotBooked,
		},
		{
			name: "past slot",
			service: NewService(
				fakeBookingRepository{hasActiveBySlotIDFn: func(_ context.Context, _ uuid.UUID) (bool, error) { return false, nil }},
				fakeSlotRepository{getByIDFn: func(_ context.Context, _ uuid.UUID) (domain.Slot, bool, error) { return pastSlot, true, nil }},
				fakeTxManager{},
				fakeIDGenerator{next: uuid.New()},
				fakeClock{now: now},
				nil,
			),
			input: CreateInput{Actor: shared.Actor{UserID: uuid.New(), Role: domain.RoleUser}, SlotID: uuid.New()},
			want:  domain.ErrPastSlotBooking,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, gotErr := tt.service.Create(context.Background(), tt.input)
			if !errors.Is(gotErr, tt.want) {
				t.Fatalf("Create() error = %v, want %v", gotErr, tt.want)
			}
		})
	}
}

func TestService_Cancel_UpdatesActiveBookingAndIsIdempotent(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	bookingID := uuid.New()
	active, err := domain.NewActiveBooking(bookingID, uuid.New(), userID, nil, time.Date(2026, time.March, 24, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("NewActiveBooking() error = %v", err)
	}
	cancelled := active
	cancelled.Cancel()

	t.Run("active booking updated", func(t *testing.T) {
		t.Parallel()
		updated := false
		service := NewService(
			fakeBookingRepository{
				getByIDFn: func(_ context.Context, _ uuid.UUID) (domain.Booking, bool, error) { return active, true, nil },
				updateFn: func(_ context.Context, booking domain.Booking) error {
					updated = true
					if booking.Status != domain.BookingStatusCancelled {
						t.Fatalf("updated booking status = %q, want %q", booking.Status, domain.BookingStatusCancelled)
					}
					return nil
				},
			},
			fakeSlotRepository{},
			fakeTxManager{},
			fakeIDGenerator{next: uuid.New()},
			fakeClock{},
			nil,
		)

		out, cancelErr := service.Cancel(context.Background(), CancelInput{Actor: shared.Actor{UserID: userID, Role: domain.RoleUser}, BookingID: bookingID})
		if cancelErr != nil {
			t.Fatalf("Cancel() error = %v", cancelErr)
		}

		if out.Booking.Status != domain.BookingStatusCancelled {
			t.Fatalf("out.Booking.Status = %q, want %q", out.Booking.Status, domain.BookingStatusCancelled)
		}

		if !updated {
			t.Fatal("expected Update() to be called")
		}
	})

	t.Run("already cancelled booking not updated", func(t *testing.T) {
		t.Parallel()
		updated := false
		service := NewService(
			fakeBookingRepository{
				getByIDFn: func(_ context.Context, _ uuid.UUID) (domain.Booking, bool, error) { return cancelled, true, nil },
				updateFn: func(_ context.Context, _ domain.Booking) error {
					updated = true
					return nil
				},
			},
			fakeSlotRepository{},
			fakeTxManager{},
			fakeIDGenerator{next: uuid.New()},
			fakeClock{},
			nil,
		)

		out, cancelErr := service.Cancel(context.Background(), CancelInput{Actor: shared.Actor{UserID: userID, Role: domain.RoleUser}, BookingID: bookingID})
		if cancelErr != nil {
			t.Fatalf("Cancel() error = %v", cancelErr)
		}

		if out.Booking.Status != domain.BookingStatusCancelled {
			t.Fatalf("out.Booking.Status = %q, want %q", out.Booking.Status, domain.BookingStatusCancelled)
		}

		if updated {
			t.Fatal("Update() should not be called for already cancelled booking")
		}
	})
}

func TestService_Cancel_ReturnsExpectedErrors(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	bookingID := uuid.New()
	booking, err := domain.NewActiveBooking(bookingID, uuid.New(), uuid.New(), nil, time.Date(2026, time.March, 24, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("NewActiveBooking() error = %v", err)
	}

	tests := []struct {
		name    string
		service *Service
		input   CancelInput
		want    error
	}{
		{name: "forbidden role", service: NewService(fakeBookingRepository{}, fakeSlotRepository{}, fakeTxManager{}, fakeIDGenerator{next: uuid.New()}, fakeClock{}, nil), input: CancelInput{Actor: shared.Actor{UserID: userID, Role: domain.RoleAdmin}, BookingID: bookingID}, want: shared.ErrForbidden},
		{name: "not found", service: NewService(fakeBookingRepository{getByIDFn: func(_ context.Context, _ uuid.UUID) (domain.Booking, bool, error) {
			return domain.Booking{}, false, nil
		}}, fakeSlotRepository{}, fakeTxManager{}, fakeIDGenerator{next: uuid.New()}, fakeClock{}, nil), input: CancelInput{Actor: shared.Actor{UserID: userID, Role: domain.RoleUser}, BookingID: bookingID}, want: shared.ErrBookingNotFound},
		{name: "wrong owner", service: NewService(fakeBookingRepository{getByIDFn: func(_ context.Context, _ uuid.UUID) (domain.Booking, bool, error) { return booking, true, nil }}, fakeSlotRepository{}, fakeTxManager{}, fakeIDGenerator{next: uuid.New()}, fakeClock{}, nil), input: CancelInput{Actor: shared.Actor{UserID: userID, Role: domain.RoleUser}, BookingID: bookingID}, want: shared.ErrForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, gotErr := tt.service.Cancel(context.Background(), tt.input)
			if !errors.Is(gotErr, tt.want) {
				t.Fatalf("Cancel() error = %v, want %v", gotErr, tt.want)
			}
		})
	}
}

func TestService_ListMine_AndList(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	wantBookings := []domain.Booking{{ID: uuid.New()}, {ID: uuid.New()}}
	service := NewService(
		fakeBookingRepository{
			listByUserFutureFn: func(_ context.Context, gotUserID uuid.UUID, _ time.Time) ([]domain.Booking, error) {
				if gotUserID != userID {
					t.Fatalf("userID = %v, want %v", gotUserID, userID)
				}
				return wantBookings, nil
			},
			listFn: func(_ context.Context, page, pageSize int) ([]domain.Booking, int, error) {
				if page != 2 || pageSize != 5 {
					t.Fatalf("page/pageSize = %d/%d, want 2/5", page, pageSize)
				}
				return wantBookings, 17, nil
			},
		},
		fakeSlotRepository{},
		fakeTxManager{},
		fakeIDGenerator{next: uuid.New()},
		fakeClock{now: time.Date(2026, time.March, 24, 12, 0, 0, 0, time.UTC)},
		nil,
	)

	mine, err := service.ListMine(context.Background(), ListMineInput{Actor: shared.Actor{UserID: userID, Role: domain.RoleUser}})
	if err != nil {
		t.Fatalf("ListMine() error = %v", err)
	}
	if len(mine.Bookings) != 2 {
		t.Fatalf("len(mine.Bookings) = %d, want 2", len(mine.Bookings))
	}

	listed, err := service.List(context.Background(), ListInput{Actor: shared.Actor{UserID: uuid.New(), Role: domain.RoleAdmin}, Page: 2, PageSize: 5})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if listed.Pagination.Total != 17 {
		t.Fatalf("listed.Pagination.Total = %d, want 17", listed.Pagination.Total)
	}
}

func TestService_List_ReturnsValidationErrors(t *testing.T) {
	t.Parallel()

	service := NewService(fakeBookingRepository{}, fakeSlotRepository{}, fakeTxManager{}, fakeIDGenerator{next: uuid.New()}, fakeClock{}, nil)

	tests := []struct {
		name  string
		input ListInput
		want  error
	}{
		{name: "forbidden", input: ListInput{Actor: shared.Actor{UserID: uuid.New(), Role: domain.RoleUser}, Page: 1, PageSize: 10}, want: shared.ErrForbidden},
		{name: "invalid page", input: ListInput{Actor: shared.Actor{UserID: uuid.New(), Role: domain.RoleAdmin}, Page: 0, PageSize: 10}, want: shared.ErrInvalidPage},
		{name: "invalid page size", input: ListInput{Actor: shared.Actor{UserID: uuid.New(), Role: domain.RoleAdmin}, Page: 1, PageSize: 0}, want: shared.ErrInvalidPageSize},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := service.List(context.Background(), tt.input)
			if !errors.Is(err, tt.want) {
				t.Fatalf("List() error = %v, want %v", err, tt.want)
			}
		})
	}
}

var _ appports.BookingRepository = fakeBookingRepository{}
var _ appports.SlotRepository = fakeSlotRepository{}
var _ appports.TxManager = fakeTxManager{}
var _ appports.Clock = fakeClock{}
var _ appports.ConferenceLinkService = fakeConferenceLinkService{}
