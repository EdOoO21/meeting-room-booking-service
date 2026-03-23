package ports

import (
	"context"

	"github.com/google/uuid"
)

type ConferenceLinkService interface {
	CreateConferenceLink(ctx context.Context, bookingID uuid.UUID) (string, error)
}
