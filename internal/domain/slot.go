package domain

import (
	"time"

	"github.com/google/uuid"
)

type Slot struct {
	ID     uuid.UUID
	RoomID uuid.UUID
	Start  time.Time
	End    time.Time
}

func NewSlot(id, roomID uuid.UUID, start, end time.Time) (Slot, error) {
	if id == uuid.Nil || roomID == uuid.Nil {
		return Slot{}, ErrInvalidID
	}

	normalizedStart, err := requireUTC(start)
	if err != nil {
		return Slot{}, err
	}

	normalizedEnd, err := requireUTC(end)
	if err != nil {
		return Slot{}, err
	}

	if !normalizedStart.Before(normalizedEnd) {
		return Slot{}, ErrSlotTimeRangeOrder
	}

	if normalizedEnd.Sub(normalizedStart) != SlotDuration {
		return Slot{}, ErrSlotDurationMustMatch
	}

	return Slot{
		ID:     id,
		RoomID: roomID,
		Start:  normalizedStart,
		End:    normalizedEnd,
	}, nil
}

func (s Slot) IsPast(now time.Time) bool {
	return !s.Start.After(normalizeUTC(now))
}

func (s Slot) IsFuture(now time.Time) bool {
	return s.Start.After(normalizeUTC(now))
}

func (s Slot) Overlaps(other Slot) bool {
	if s.RoomID != other.RoomID {
		return false
	}

	return s.Start.Before(other.End) && other.Start.Before(s.End)
}
