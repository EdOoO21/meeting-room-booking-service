package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewRoom_NormalizesFieldsAndClonesCapacity(t *testing.T) {
	t.Parallel()

	capacity := 8
	createdAt := time.Date(2026, time.March, 24, 15, 0, 0, 0, time.FixedZone("UTC+3", 3*60*60))

	room, err := NewRoom(uuid.New(), "  Focus room  ", "  Quiet place  ", &capacity, createdAt)
	if err != nil {
		t.Fatalf("NewRoom() error = %v", err)
	}

	if room.Name != "Focus room" {
		t.Fatalf("room.Name = %q, want %q", room.Name, "Focus room")
	}

	if room.Description != "Quiet place" {
		t.Fatalf("room.Description = %q, want %q", room.Description, "Quiet place")
	}

	if room.Capacity == nil || *room.Capacity != 8 {
		t.Fatalf("room.Capacity = %v, want 8", room.Capacity)
	}

	capacity = 99
	if *room.Capacity != 8 {
		t.Fatalf("room.Capacity changed after input mutation = %d, want 8", *room.Capacity)
	}

	if room.CreatedAt.Location() != time.UTC {
		t.Fatalf("room.CreatedAt location = %v, want UTC", room.CreatedAt.Location())
	}
}

func TestNewRoom_ReturnsValidationErrors(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, time.March, 24, 12, 0, 0, 0, time.UTC)
	invalidCapacity := 0

	tests := []struct {
		name     string
		id       uuid.UUID
		nameArg  string
		capacity *int
		want     error
	}{
		{name: "invalid id", id: uuid.Nil, nameArg: "Room", want: ErrInvalidID},
		{name: "empty name", id: uuid.New(), nameArg: "   ", want: ErrInvalidRoomName},
		{name: "non positive capacity", id: uuid.New(), nameArg: "Room", capacity: &invalidCapacity, want: ErrInvalidRoomCapacity},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := NewRoom(tt.id, tt.nameArg, "desc", tt.capacity, createdAt)
			if !errors.Is(err, tt.want) {
				t.Fatalf("NewRoom() error = %v, want %v", err, tt.want)
			}
		})
	}
}
