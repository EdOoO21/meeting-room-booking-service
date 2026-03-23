package ports

import "github.com/google/uuid"

// IDGenerator создает идентификаторы для новых сущностей
type IDGenerator interface {
	// NewUUID возвращает новый UUID
	NewUUID() uuid.UUID
}
