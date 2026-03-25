package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewUser_NormalizesEmailAndCreatedAt(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	createdAt := time.Date(2026, time.March, 24, 12, 0, 0, 0, time.FixedZone("UTC+3", 3*60*60))

	user, err := NewUser(id, "  USER@Example.COM  ", RoleAdmin, createdAt)
	if err != nil {
		t.Fatalf("NewUser() error = %v", err)
	}

	if user.ID != id {
		t.Fatalf("user.ID = %v, want %v", user.ID, id)
	}

	if user.Email != "user@example.com" {
		t.Fatalf("user.Email = %q, want %q", user.Email, "user@example.com")
	}

	if user.Role != RoleAdmin {
		t.Fatalf("user.Role = %q, want %q", user.Role, RoleAdmin)
	}

	if user.CreatedAt.Location() != time.UTC {
		t.Fatalf("user.CreatedAt location = %v, want UTC", user.CreatedAt.Location())
	}

	if !user.CreatedAt.Equal(createdAt.UTC()) {
		t.Fatalf("user.CreatedAt = %v, want %v", user.CreatedAt, createdAt.UTC())
	}
}

func TestNewUser_ReturnsValidationErrors(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, time.March, 24, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name  string
		id    uuid.UUID
		email string
		role  Role
		want  error
	}{
		{name: "invalid id", id: uuid.Nil, email: "user@example.com", role: RoleUser, want: ErrInvalidID},
		{name: "empty email", id: uuid.New(), email: "   ", role: RoleUser, want: ErrInvalidEmail},
		{name: "malformed email", id: uuid.New(), email: "not-an-email", role: RoleUser, want: ErrInvalidEmail},
		{name: "invalid role", id: uuid.New(), email: "user@example.com", role: Role("manager"), want: ErrInvalidRole},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := NewUser(tt.id, tt.email, tt.role, createdAt)
			if !errors.Is(err, tt.want) {
				t.Fatalf("NewUser() error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestRoleCapabilities(t *testing.T) {
	t.Parallel()

	if !RoleAdmin.IsValid() {
		t.Fatal("RoleAdmin should be valid")
	}

	if !RoleUser.IsValid() {
		t.Fatal("RoleUser should be valid")
	}

	if Role("unknown").IsValid() {
		t.Fatal("unknown role should be invalid")
	}

	if !RoleAdmin.CanManageRooms() {
		t.Fatal("admin should manage rooms")
	}

	if RoleUser.CanManageRooms() {
		t.Fatal("user should not manage rooms")
	}

	if !RoleUser.CanCreateBooking() {
		t.Fatal("user should create bookings")
	}

	if RoleAdmin.CanCreateBooking() {
		t.Fatal("admin should not create bookings")
	}
}
