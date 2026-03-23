package ports

import (
	"context"

	"github.com/google/uuid"
)

// ConferenceLinkService создает ссылку на конференцию для брони
type ConferenceLinkService interface {
	// CreateConferenceLink создает и возвращает ссылку для указанной брони
	CreateConferenceLink(ctx context.Context, bookingID uuid.UUID) (string, error)
}
