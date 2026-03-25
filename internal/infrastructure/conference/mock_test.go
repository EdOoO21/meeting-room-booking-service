package conference

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestMockService_CreateConferenceLink(t *testing.T) {
	t.Parallel()

	bookingID := uuid.New()
	_, err := NewMock().CreateConferenceLink(context.Background(), bookingID)
	if err != nil {
		t.Fatalf("CreateConferenceLink() error = %v", err)
	}
}
