package ports

import "github.com/google/uuid"

type IDGenerator interface {
	NewUUID() uuid.UUID
}
