package domain

import (
	"time"

	"github.com/google/uuid"
)

type BookingStatus string

const (
	BookingStatusActive    BookingStatus = "active"
	BookingStatusCancelled BookingStatus = "cancelled"
)

type Booking struct {
	ID             uuid.UUID
	SlotID         uuid.UUID
	UserID         uuid.UUID
	Status         BookingStatus
	ConferenceLink *string
	CreatedAt      time.Time
}

func NewBooking(id, slotID, userID uuid.UUID, status BookingStatus, conferenceLink *string, createdAt time.Time) (Booking, error) {
	if id == uuid.Nil || slotID == uuid.Nil || userID == uuid.Nil {
		return Booking{}, ErrInvalidID
	}

	if !status.IsValid() {
		return Booking{}, ErrInvalidBookingStatus
	}

	return Booking{
		ID:             id,
		SlotID:         slotID,
		UserID:         userID,
		Status:         status,
		ConferenceLink: cloneStringPointer(conferenceLink),
		CreatedAt:      normalizeUTC(createdAt),
	}, nil
}

func NewActiveBooking(id, slotID, userID uuid.UUID, conferenceLink *string, createdAt time.Time) (Booking, error) {
	return NewBooking(id, slotID, userID, BookingStatusActive, conferenceLink, createdAt)
}

func (s BookingStatus) IsValid() bool {
	return s == BookingStatusActive || s == BookingStatusCancelled
}

func (b *Booking) BelongsTo(userID uuid.UUID) bool {
	return b.UserID == userID
}

func (b *Booking) IsActive() bool {
	return b.Status == BookingStatusActive
}

func (b *Booking) Cancel() {
	b.Status = BookingStatusCancelled
}

func (b *Booking) CanBeCreatedBy(role Role, slot Slot, now time.Time) error {
	if !role.CanCreateBooking() {
		return ErrForbiddenBookingByRole
	}

	if slot.IsPast(now) {
		return ErrPastSlotBooking
	}

	return nil
}

func cloneStringPointer(value *string) *string {
	if value == nil {
		return nil
	}

	cloned := *value
	return &cloned
}
