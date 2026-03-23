package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewBooking_ClonesConferenceLinkAndNormalizesTime(t *testing.T) {
	t.Parallel()

	link := "https://meet.example/room"
	createdAt := time.Date(2026, time.March, 24, 12, 0, 0, 0, time.FixedZone("UTC+3", 3*60*60))

	booking, err := NewBooking(uuid.New(), uuid.New(), uuid.New(), BookingStatusActive, &link, createdAt)
	if err != nil {
		t.Fatalf("NewBooking() error = %v", err)
	}

	if booking.ConferenceLink == nil || *booking.ConferenceLink != link {
		t.Fatalf("booking.ConferenceLink = %v, want %q", booking.ConferenceLink, link)
	}

	link = "changed"
	if *booking.ConferenceLink != "https://meet.example/room" {
		t.Fatalf("booking.ConferenceLink changed after input mutation = %q", *booking.ConferenceLink)
	}

	if booking.CreatedAt.Location() != time.UTC {
		t.Fatalf("booking.CreatedAt location = %v, want UTC", booking.CreatedAt.Location())
	}
}

func TestNewBooking_ReturnsValidationErrors(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, time.March, 24, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name   string
		id     uuid.UUID
		slotID uuid.UUID
		userID uuid.UUID
		status BookingStatus
		want   error
	}{
		{name: "invalid id", id: uuid.Nil, slotID: uuid.New(), userID: uuid.New(), status: BookingStatusActive, want: ErrInvalidID},
		{name: "invalid slot id", id: uuid.New(), slotID: uuid.Nil, userID: uuid.New(), status: BookingStatusActive, want: ErrInvalidID},
		{name: "invalid user id", id: uuid.New(), slotID: uuid.New(), userID: uuid.Nil, status: BookingStatusActive, want: ErrInvalidID},
		{name: "invalid status", id: uuid.New(), slotID: uuid.New(), userID: uuid.New(), status: BookingStatus("done"), want: ErrInvalidBookingStatus},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := NewBooking(tt.id, tt.slotID, tt.userID, tt.status, nil, createdAt)
			if !errors.Is(err, tt.want) {
				t.Fatalf("NewBooking() error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestBookingBehavior(t *testing.T) {
	t.Parallel()

	booking, err := NewActiveBooking(uuid.New(), uuid.New(), uuid.New(), nil, time.Date(2026, time.March, 24, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("NewActiveBooking() error = %v", err)
	}

	if !booking.IsActive() {
		t.Fatal("new active booking should be active")
	}

	if !booking.BelongsTo(booking.UserID) {
		t.Fatal("booking should belong to its user")
	}

	booking.Cancel()
	if booking.Status != BookingStatusCancelled {
		t.Fatalf("booking.Status = %q, want %q", booking.Status, BookingStatusCancelled)
	}
}

func TestBookingCanBeCreatedBy(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.March, 24, 12, 0, 0, 0, time.UTC)
	futureSlot, err := NewSlot(uuid.New(), uuid.New(), now.Add(time.Hour), now.Add(time.Hour).Add(SlotDuration))
	if err != nil {
		t.Fatalf("NewSlot() future error = %v", err)
	}

	pastSlot, err := NewSlot(uuid.New(), uuid.New(), now.Add(-time.Hour), now.Add(-time.Hour).Add(SlotDuration))
	if err != nil {
		t.Fatalf("NewSlot() past error = %v", err)
	}

	booking := Booking{}

	if err := booking.CanBeCreatedBy(RoleUser, futureSlot, now); err != nil {
		t.Fatalf("CanBeCreatedBy() unexpected error = %v", err)
	}

	if err := booking.CanBeCreatedBy(RoleAdmin, futureSlot, now); !errors.Is(err, ErrForbiddenBookingByRole) {
		t.Fatalf("CanBeCreatedBy() error = %v, want %v", err, ErrForbiddenBookingByRole)
	}

	if err := booking.CanBeCreatedBy(RoleUser, pastSlot, now); !errors.Is(err, ErrPastSlotBooking) {
		t.Fatalf("CanBeCreatedBy() error = %v, want %v", err, ErrPastSlotBooking)
	}
}
