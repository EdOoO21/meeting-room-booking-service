package id

import (
	"testing"

	"github.com/google/uuid"
)

func TestGenerator_NewUUID_ReturnsNonNilUUID(t *testing.T) {
	t.Parallel()

	got := New().NewUUID()
	if got == uuid.Nil {
		t.Fatal("expected non-nil UUID")
	}
}
