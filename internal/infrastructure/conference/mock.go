package conference

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type MockService struct{}

func NewMock() MockService {
	return MockService{}
}

func (MockService) CreateConferenceLink(ctx context.Context, bookingID uuid.UUID) (string, error) {
	_ = ctx
	return fmt.Sprintf("https://conference.local/rooms/%s", bookingID.String()), nil
}
