package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewSlot_ReturnsNormalizedUTCSlot(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, time.March, 25, 9, 0, 0, 0, time.UTC)
	end := start.Add(SlotDuration)

	slot, err := NewSlot(uuid.New(), uuid.New(), start, end)
	if err != nil {
		t.Fatalf("NewSlot() error = %v", err)
	}

	if !slot.Start.Equal(start) || !slot.End.Equal(end) {
		t.Fatalf("slot times = %v..%v, want %v..%v", slot.Start, slot.End, start, end)
	}
}

func TestNewSlot_ReturnsValidationErrors(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, time.March, 25, 9, 0, 0, 0, time.UTC)
	nonUTCStart := time.Date(2026, time.March, 25, 9, 0, 0, 0, time.FixedZone("UTC+3", 3*60*60))

	tests := []struct {
		name   string
		id     uuid.UUID
		roomID uuid.UUID
		start  time.Time
		end    time.Time
		want   error
	}{
		{name: "invalid id", id: uuid.Nil, roomID: uuid.New(), start: start, end: start.Add(SlotDuration), want: ErrInvalidID},
		{name: "invalid room id", id: uuid.New(), roomID: uuid.Nil, start: start, end: start.Add(SlotDuration), want: ErrInvalidID},
		{name: "zero start", id: uuid.New(), roomID: uuid.New(), start: time.Time{}, end: start.Add(SlotDuration), want: ErrInvalidTimestamp},
		{name: "non utc start", id: uuid.New(), roomID: uuid.New(), start: nonUTCStart, end: nonUTCStart.Add(SlotDuration), want: ErrNonUTCTimestamp},
		{name: "start after end", id: uuid.New(), roomID: uuid.New(), start: start, end: start, want: ErrSlotTimeRangeOrder},
		{name: "duration mismatch", id: uuid.New(), roomID: uuid.New(), start: start, end: start.Add(45 * time.Minute), want: ErrSlotDurationMustMatch},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := NewSlot(tt.id, tt.roomID, tt.start, tt.end)
			if !errors.Is(err, tt.want) {
				t.Fatalf("NewSlot() error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestSlotTimeHelpersAndOverlap(t *testing.T) {
	t.Parallel()

	roomID := uuid.New()
	base := time.Date(2026, time.March, 25, 10, 0, 0, 0, time.UTC)

	slot, err := NewSlot(uuid.New(), roomID, base, base.Add(SlotDuration))
	if err != nil {
		t.Fatalf("NewSlot() error = %v", err)
	}

	if !slot.IsPast(base.Add(5 * time.Minute)) {
		t.Fatal("slot should be past after it starts")
	}

	if !slot.IsFuture(base.Add(-time.Minute)) {
		t.Fatal("slot should be future before it starts")
	}

	overlapping, err := NewSlot(uuid.New(), roomID, base.Add(15*time.Minute), base.Add(45*time.Minute))
	if err != nil {
		t.Fatalf("NewSlot() overlapping error = %v", err)
	}

	otherRoom, err := NewSlot(uuid.New(), uuid.New(), base.Add(15*time.Minute), base.Add(45*time.Minute))
	if err != nil {
		t.Fatalf("NewSlot() otherRoom error = %v", err)
	}

	if !slot.Overlaps(overlapping) {
		t.Fatal("slots in same room should overlap")
	}

	if slot.Overlaps(otherRoom) {
		t.Fatal("slots in different rooms should not overlap")
	}
}
