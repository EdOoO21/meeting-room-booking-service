package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type Room struct {
	ID          uuid.UUID
	Name        string
	Description string
	Capacity    *int
	CreatedAt   time.Time
}

func NewRoom(id uuid.UUID, name, description string, capacity *int, createdAt time.Time) (Room, error) {
	if id == uuid.Nil {
		return Room{}, ErrInvalidID
	}

	normalizedName := strings.TrimSpace(name)
	if normalizedName == "" {
		return Room{}, ErrInvalidRoomName
	}

	normalizedDescription := strings.TrimSpace(description)

	if capacity != nil && *capacity <= 0 {
		return Room{}, ErrInvalidRoomCapacity
	}

	return Room{
		ID:          id,
		Name:        normalizedName,
		Description: normalizedDescription,
		Capacity:    cloneIntPointer(capacity),
		CreatedAt:   normalizeUTC(createdAt),
	}, nil
}

func cloneIntPointer(value *int) *int {
	if value == nil {
		return nil
	}

	cloned := *value
	return &cloned
}
