package domain

import (
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

type User struct {
	ID        uuid.UUID
	Email     string
	Role      Role
	CreatedAt time.Time
}

func NewUser(id uuid.UUID, email string, role Role, createdAt time.Time) (User, error) {
	if id == uuid.Nil {
		return User{}, ErrInvalidID
	}

	normalizedEmail := strings.TrimSpace(strings.ToLower(email))
	if normalizedEmail == "" {
		return User{}, ErrInvalidEmail
	}

	if _, err := mail.ParseAddress(normalizedEmail); err != nil {
		return User{}, fmt.Errorf("%w: %s", ErrInvalidEmail, err.Error())
	}

	if !role.IsValid() {
		return User{}, ErrInvalidRole
	}

	return User{
		ID:        id,
		Email:     normalizedEmail,
		Role:      role,
		CreatedAt: normalizeUTC(createdAt),
	}, nil
}

func (r Role) IsValid() bool {
	return r == RoleAdmin || r == RoleUser
}

func (r Role) CanManageRooms() bool {
	return r == RoleAdmin
}

func (r Role) CanCreateBooking() bool {
	return r == RoleUser
}
